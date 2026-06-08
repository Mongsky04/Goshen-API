// goshen-api/src/routes/brands.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { brands } from '../db/schema.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { uploadIfPresent, MimeError } from '../lib/route-helpers.js'

export const brandRoutes = new Hono()

brandRoutes.get('/', async (c) => {
  try {
    const rows = await db.select().from(brands).orderBy(asc(brands.sortOrder), asc(brands.id))
    return ok(c, rows)
  } catch {
    return internalError(c)
  }
})

brandRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [row] = await db.select().from(brands).where(eq(brands.id, id)).limit(1)
    if (!row) return notFound(c, 'brand not found')
    return ok(c, row)
  } catch {
    return internalError(c)
  }
})

brandRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const name = form.get('name') as string
    if (!name) return badRequest(c, 'name is required')
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? ''
    const [row] = await db.insert(brands).values({ name, imageUrl }).returning()
    return created(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

brandRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [existing] = await db.select().from(brands).where(eq(brands.id, id)).limit(1)
    if (!existing) return notFound(c, 'brand not found')
    const form = await c.req.formData()
    const name = (form.get('name') as string) || existing.name
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? existing.imageUrl
    const [row] = await db.update(brands).set({ name, imageUrl }).where(eq(brands.id, id)).returning()
    return ok(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

brandRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    await db.delete(brands).where(eq(brands.id, id))
    return ok(c, null)
  } catch {
    return internalError(c)
  }
})
