// goshen-api/src/routes/featured.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { featured, products } from '../db/schema.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { parsePage } from '../lib/route-helpers.js'

const featuredCols = {
  id: featured.id,
  productId: featured.productId,
  featuredCategories: featured.featuredCategories,
  sortOrder: featured.sortOrder,
  createdAt: featured.createdAt,
  updatedAt: featured.updatedAt,
  name: products.name,
  imageUrl: products.imageUrl,
  category: products.category,
  subCategory: products.subCategory,
}

async function getFeaturedById(id: number) {
  const rows = await db.select(featuredCols).from(featured)
    .innerJoin(products, eq(featured.productId, products.id))
    .where(eq(featured.id, id)).limit(1)
  return rows[0] ?? null
}

export const featuredRoutes = new Hono()

featuredRoutes.get('/', async (c) => {
  try {
    const { page, limit, offset } = parsePage(c.req.query.bind(c.req))
    const rows = await db.select(featuredCols).from(featured)
      .innerJoin(products, eq(featured.productId, products.id))
      .orderBy(asc(featured.sortOrder), asc(featured.id))
      .limit(limit).offset(offset)
    return ok(c, { data: rows, page, limit, total: null })
  } catch {
    return internalError(c)
  }
})

featuredRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const row = await getFeaturedById(id)
    if (!row) return notFound(c, 'featured not found')
    return ok(c, row)
  } catch {
    return internalError(c)
  }
})

featuredRoutes.post('/', requireAuth, async (c) => {
  try {
    const body = await c.req.json<{ product_id?: number; featured_categories?: string[] }>()
    if (!body.product_id) return badRequest(c, 'product_id is required')
    const [product] = await db.select().from(products).where(eq(products.id, body.product_id)).limit(1)
    if (!product) return notFound(c, 'product not found')
    const [row] = await db.insert(featured).values({
      productId: product.id,
      name: product.name,
      imageUrl: product.imageUrl,
      category: product.category,
      subCategory: product.subCategory,
      featuredCategories: body.featured_categories ?? [],
    }).returning()
    return created(c, await getFeaturedById(row.id))
  } catch {
    return internalError(c)
  }
})

featuredRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const existing = await getFeaturedById(id)
    if (!existing) return notFound(c, 'featured not found')
    const body = await c.req.json<{ product_id?: number; featured_categories?: string[] }>()
    const nextProductId = body.product_id ?? existing.productId
    const [product] = await db.select().from(products).where(eq(products.id, nextProductId)).limit(1)
    if (!product) return notFound(c, 'product not found')
    await db.update(featured).set({
      productId: product.id,
      name: product.name,
      imageUrl: product.imageUrl,
      category: product.category,
      subCategory: product.subCategory,
      featuredCategories: body.featured_categories ?? existing.featuredCategories,
    }).where(eq(featured.id, id))
    return ok(c, await getFeaturedById(id))
  } catch {
    return internalError(c)
  }
})

featuredRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    await db.delete(featured).where(eq(featured.id, id))
    return ok(c, null)
  } catch {
    return internalError(c)
  }
})
