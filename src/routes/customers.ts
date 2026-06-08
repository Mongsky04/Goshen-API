// goshen-api/src/routes/customers.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { customers } from '../db/schema.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { uploadIfPresent, MimeError } from '../lib/route-helpers.js'

export const customerRoutes = new Hono()

customerRoutes.get('/', async (c) => {
  try {
    const rows = await db.select().from(customers).orderBy(asc(customers.sortOrder), asc(customers.id))
    return ok(c, rows)
  } catch {
    return internalError(c)
  }
})

customerRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [row] = await db.select().from(customers).where(eq(customers.id, id)).limit(1)
    if (!row) return notFound(c, 'customer not found')
    return ok(c, row)
  } catch {
    return internalError(c)
  }
})

customerRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const altText = (form.get('alt_text') as string) ?? ''
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? ''
    const [row] = await db.insert(customers).values({ altText, imageUrl }).returning()
    return created(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

customerRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [existing] = await db.select().from(customers).where(eq(customers.id, id)).limit(1)
    if (!existing) return notFound(c, 'customer not found')
    const form = await c.req.formData()
    const altText = (form.get('alt_text') as string) ?? existing.altText
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? existing.imageUrl
    const [row] = await db.update(customers).set({ altText, imageUrl }).where(eq(customers.id, id)).returning()
    return ok(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

customerRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    await db.delete(customers).where(eq(customers.id, id))
    return ok(c, null)
  } catch {
    return internalError(c)
  }
})
