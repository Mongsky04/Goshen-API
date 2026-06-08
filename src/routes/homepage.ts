// goshen-api/src/routes/homepage.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import {
  homepageSupportCards, homepageGridProducts, products,
  pageBanners, slider, conferencePages, performerPages,
} from '../db/schema.js'
import { ok, created, badRequest, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'

export const homepageRoutes = new Hono()

// GET /api/v1/support-cards
homepageRoutes.get('/support-cards', async (c) => {
  try {
    const rows = await db.select().from(homepageSupportCards).orderBy(asc(homepageSupportCards.sortOrder))
    return ok(c, rows)
  } catch { return internalError(c) }
})

// POST /api/v1/support-cards (protected)
homepageRoutes.post('/support-cards', requireAuth, async (c) => {
  try {
    const body = await c.req.json<{ title?: string; description?: string; cta_label?: string; cta_href?: string }>()
    const [row] = await db.insert(homepageSupportCards).values({
      title: body.title ?? '',
      description: body.description ?? '',
      ctaLabel: body.cta_label ?? '',
      ctaHref: body.cta_href ?? '#',
    }).returning()
    return created(c, row)
  } catch { return internalError(c) }
})

// DELETE /api/v1/support-cards/:id (protected)
homepageRoutes.delete('/support-cards/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    await db.delete(homepageSupportCards).where(eq(homepageSupportCards.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})

// GET /api/v1/homepage-grid
homepageRoutes.get('/homepage-grid', async (c) => {
  try {
    const rows = await db
      .select({
        id: homepageGridProducts.id,
        productId: homepageGridProducts.productId,
        name: products.name,
        imageUrl: products.imageUrl,
        category: products.category,
        subCategory: products.subCategory,
      })
      .from(homepageGridProducts)
      .innerJoin(products, eq(homepageGridProducts.productId, products.id))
      .orderBy(asc(homepageGridProducts.sortOrder))
    return ok(c, rows)
  } catch { return internalError(c) }
})

// PUT /api/v1/homepage-grid (protected) — replaces all rows
homepageRoutes.put('/homepage-grid', requireAuth, async (c) => {
  try {
    const body = await c.req.json<{ product_ids?: number[] }>()
    const ids = body.product_ids ?? []
    await db.delete(homepageGridProducts)
    if (ids.length > 0) {
      await db.insert(homepageGridProducts).values(
        ids.map((productId, i) => ({ productId, sortOrder: i }))
      )
    }
    return ok(c, null)
  } catch { return internalError(c) }
})

// GET /api/v1/nav
// NOTE: returns { hidden: string[] } directly, NOT the standard { success, data } wrapper
// This matches the Go implementation which used json.NewEncoder directly
homepageRoutes.get('/nav', async (c) => {
  try {
    const confHidden = await db
      .select({ slug: conferencePages.slug })
      .from(conferencePages)
      .where(eq(conferencePages.isPublished, false))
    const perfHidden = await db
      .select({ slug: performerPages.slug })
      .from(performerPages)
      .where(eq(performerPages.isPublished, false))
    const hidden = [...confHidden.map(r => r.slug), ...perfHidden.map(r => r.slug)]
    c.header('Cache-Control', 'public, max-age=60')
    return c.json({ hidden })
  } catch { return internalError(c) }
})

// GET /api/v1/page-banners/:slug
homepageRoutes.get('/page-banners/:slug', async (c) => {
  try {
    const slug = c.req.param('slug')
    const rows = await db
      .select({
        id: slider.id,
        title: slider.title,
        imageUrl: slider.imageUrl,
        orderNum: pageBanners.orderNum,
      })
      .from(pageBanners)
      .innerJoin(slider, eq(pageBanners.bannerId, slider.id))
      .where(eq(pageBanners.pageSlug, slug))
      .orderBy(asc(pageBanners.orderNum))
    return ok(c, rows)
  } catch { return internalError(c) }
})

// PUT /api/v1/page-banners/:slug (protected) — replaces banners for a slug
homepageRoutes.put('/page-banners/:slug', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    const body = await c.req.json<{ banner_ids?: number[] }>()
    const ids = body.banner_ids ?? []
    await db.delete(pageBanners).where(eq(pageBanners.pageSlug, slug))
    if (ids.length > 0) {
      await db.insert(pageBanners).values(
        ids.map((bannerId, i) => ({ pageSlug: slug, bannerId, orderNum: i }))
      )
    }
    return ok(c, null)
  } catch { return internalError(c) }
})
