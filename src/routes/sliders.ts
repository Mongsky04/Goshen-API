// goshen-api/src/routes/sliders.ts
// NOTE: mounted at /api/v1/banners, uses the `slider` table
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { slider } from '../db/schema.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { uploadIfPresent, MimeError } from '../lib/route-helpers.js'
import { deleteCloudinaryAsset } from '../lib/storage.js'

export const sliderRoutes = new Hono()

sliderRoutes.get('/', async (c) => {
  try {
    const rows = await db.select().from(slider).orderBy(asc(slider.orderNum), asc(slider.id))
    return ok(c, rows)
  } catch {
    return internalError(c)
  }
})

sliderRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [row] = await db.select().from(slider).where(eq(slider.id, id)).limit(1)
    if (!row) return notFound(c, 'slider not found')
    return ok(c, row)
  } catch {
    return internalError(c)
  }
})

sliderRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const title = (form.get('title') as string) ?? ''
    const orderNum = parseInt((form.get('order_num') as string) ?? '0') || 0
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? ''
    const [row] = await db.insert(slider).values({ title, imageUrl, orderNum }).returning()
    return created(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

sliderRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [existing] = await db.select().from(slider).where(eq(slider.id, id)).limit(1)
    if (!existing) return notFound(c, 'slider not found')
    const form = await c.req.formData()
    const title = (form.get('title') as string) ?? existing.title
    const orderNum = parseInt((form.get('order_num') as string) ?? '') || existing.orderNum
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? existing.imageUrl
    const [row] = await db.update(slider).set({ title, imageUrl, orderNum }).where(eq(slider.id, id)).returning()
    return ok(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

sliderRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [deleted] = await db.delete(slider).where(eq(slider.id, id)).returning()
    if (deleted?.imageUrl) await deleteCloudinaryAsset(deleted.imageUrl)
    return ok(c, null)
  } catch {
    return internalError(c)
  }
})
