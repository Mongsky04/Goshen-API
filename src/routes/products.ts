// goshen-api/src/routes/products.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { products } from '../db/schema.js'
import { deleteCloudinaryAsset } from '../lib/storage.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { parsePage, uploadIfPresent, MimeError } from '../lib/route-helpers.js'

export const productRoutes = new Hono()

productRoutes.get('/', async (c) => {
  try {
    const { page, limit, offset } = parsePage(c.req.query.bind(c.req))
    const rows = await db.select().from(products)
      .orderBy(asc(products.sortOrder), asc(products.id))
      .limit(limit).offset(offset)
    return ok(c, { data: rows, page, limit, total: null })
  } catch {
    return internalError(c)
  }
})

productRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [row] = await db.select().from(products).where(eq(products.id, id)).limit(1)
    if (!row) return notFound(c, 'product not found')
    return ok(c, row)
  } catch {
    return internalError(c)
  }
})

productRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const name = form.get('name') as string
    if (!name) return badRequest(c, 'name is required')
    const category = (form.get('category') as string) ?? ''
    const subCategory = (form.get('sub_category') as string) ?? ''
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? ''
    const [row] = await db.insert(products).values({ name, imageUrl, category, subCategory }).returning()
    return created(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

productRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [existing] = await db.select().from(products).where(eq(products.id, id)).limit(1)
    if (!existing) return notFound(c, 'product not found')
    const form = await c.req.formData()
    const name = (form.get('name') as string) || existing.name
    const category = (form.get('category') as string) ?? existing.category
    const subCategory = (form.get('sub_category') as string) ?? existing.subCategory
    const imageUrl = (await uploadIfPresent(form, 'image')) ?? existing.imageUrl
    const [row] = await db.update(products).set({ name, imageUrl, category, subCategory }).where(eq(products.id, id)).returning()
    return ok(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

productRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [deleted] = await db.delete(products).where(eq(products.id, id)).returning()
    if (deleted?.imageUrl) await deleteCloudinaryAsset(deleted.imageUrl)
    return ok(c, null)
  } catch {
    return internalError(c)
  }
})
