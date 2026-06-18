// goshen-api/src/routes/performer.ts
import { Hono } from 'hono'
import { and, eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { performerPages, performerProducts, performerVideos, products } from '../db/schema.js'
import { ok, created, notFound, internalError, badRequest } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'

const VALID_SLUGS = new Set(['musician', 'vocalist', 'master-ceremony'])

interface PerformerPutBody {
  isPublished?: boolean
  heroImageUrl?: string
  productGridTitle?: string
  videosSectionTitle?: string
  mainVideo?: {
    title?: string
    subtitle?: string
    thumbnailUrl?: string
    videoUrl?: string
  }
}

async function buildPerformerPage(slug: string) {
  const [page] = await db.select().from(performerPages).where(eq(performerPages.slug, slug)).limit(1)
  if (!page) return null

  const productRows = await db
    .select({
      id: performerProducts.id,
      productId: performerProducts.productId,
      isHidden: performerProducts.isHidden,
      sortOrder: performerProducts.sortOrder,
      name: products.name,
      category: products.category,
      subCategory: products.subCategory,
      imageUrl: products.imageUrl,
    })
    .from(performerProducts)
    .innerJoin(products, eq(performerProducts.productId, products.id))
    .where(eq(performerProducts.pageId, page.id))
    .orderBy(asc(performerProducts.sortOrder))

  const videos = await db.select().from(performerVideos)
    .where(eq(performerVideos.pageId, page.id))
    .orderBy(asc(performerVideos.sortOrder))

  const mainVideo = videos.find(v => v.isMain) ?? null
  const relatedVideos = videos.filter(v => !v.isMain)

  return {
    id: page.id,
    slug: page.slug,
    label: page.label,
    isPublished: page.isPublished,
    heroImageUrl: page.heroImageUrl,
    productGridTitle: page.productGridTitle,
    videosSectionTitle: page.videosSectionTitle,
    products: productRows,
    mainVideo,
    relatedVideos,
  }
}

export const performerRoutes = new Hono()

// GET /api/v1/performer-pages/:slug (public)
performerRoutes.get('/performer-pages/:slug', async (c) => {
  try {
    const data = await buildPerformerPage(c.req.param('slug'))
    if (!data) return notFound(c, 'performer page not found')
    return ok(c, data)
  } catch { return internalError(c) }
})

// GET /api/v1/admin/performer-pages/:slug (protected)
performerRoutes.get('/admin/performer-pages/:slug', requireAuth, async (c) => {
  try {
    const data = await buildPerformerPage(c.req.param('slug'))
    if (!data) return notFound(c, 'performer page not found')
    return ok(c, data)
  } catch { return internalError(c) }
})

// PUT /api/v1/admin/performer-pages/:slug (protected)
// Accepts JSON body: { isPublished, heroImageUrl, productGridTitle, videosSectionTitle, mainVideo }
performerRoutes.put('/admin/performer-pages/:slug', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    if (!VALID_SLUGS.has(slug)) return badRequest(c, 'invalid performer slug')

    const body = await c.req.json<PerformerPutBody>()

    const [page] = await db.select().from(performerPages).where(eq(performerPages.slug, slug)).limit(1)
    if (!page) return notFound(c, 'performer page not found')

    await db.update(performerPages).set({
      heroImageUrl: body.heroImageUrl ?? page.heroImageUrl,
      isPublished: body.isPublished ?? page.isPublished,
      productGridTitle: body.productGridTitle ?? page.productGridTitle,
      videosSectionTitle: body.videosSectionTitle ?? page.videosSectionTitle,
    }).where(eq(performerPages.id, page.id))

    if (body.mainVideo !== undefined) {
      const mv = body.mainVideo
      // Replace existing main video (delete then insert to avoid id drift)
      await db.delete(performerVideos).where(
        and(eq(performerVideos.pageId, page.id), eq(performerVideos.isMain, true))
      )
      await db.insert(performerVideos).values({
        pageId: page.id,
        isMain: true,
        title: mv.title ?? '',
        subtitle: mv.subtitle ?? '',
        thumbnailUrl: mv.thumbnailUrl ?? '',
        videoUrl: mv.videoUrl ?? '',
        sortOrder: 0,
      })
    }

    return ok(c, await buildPerformerPage(slug))
  } catch {
    return internalError(c)
  }
})

// POST /admin/performer-pages/:slug/products
performerRoutes.post('/admin/performer-pages/:slug/products', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    const [page] = await db.select().from(performerPages).where(eq(performerPages.slug, slug)).limit(1)
    if (!page) return notFound(c, 'performer page not found')
    const { productId } = await c.req.json<{ productId: number }>()
    if (!productId) return badRequest(c, 'productId required')
    const count = await db.select().from(performerProducts)
      .where(eq(performerProducts.pageId, page.id)).then(r => r.length)
    const [row] = await db.insert(performerProducts)
      .values({ pageId: page.id, productId, sortOrder: count })
      .returning()
    return created(c, row)
  } catch (e: any) {
    if (e?.code === '23505') return c.json({ success: false, error: 'product already on this page' }, 409)
    return internalError(c)
  }
})

// PATCH /admin/performer-pages/:slug/products/:id
performerRoutes.patch('/admin/performer-pages/:slug/products/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    const { isHidden } = await c.req.json<{ isHidden: boolean }>()
    const [row] = await db.update(performerProducts).set({ isHidden }).where(eq(performerProducts.id, id)).returning()
    if (!row) return notFound(c, 'not found')
    return ok(c, row)
  } catch { return internalError(c) }
})

// DELETE /admin/performer-pages/:slug/products/:id
performerRoutes.delete('/admin/performer-pages/:slug/products/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    await db.delete(performerProducts).where(eq(performerProducts.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})

// POST /admin/performer-pages/:slug/videos
performerRoutes.post('/admin/performer-pages/:slug/videos', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    const [page] = await db.select().from(performerPages).where(eq(performerPages.slug, slug)).limit(1)
    if (!page) return notFound(c, 'performer page not found')
    const { title, subtitle, thumbnailUrl, videoUrl } = await c.req.json<{
      title: string; subtitle: string; thumbnailUrl: string; videoUrl: string
    }>()
    const count = await db.select().from(performerVideos)
      .where(eq(performerVideos.pageId, page.id)).then(r => r.length)
    const [row] = await db.insert(performerVideos)
      .values({ pageId: page.id, isMain: false, title, subtitle, thumbnailUrl, videoUrl, sortOrder: count })
      .returning()
    return created(c, row)
  } catch { return internalError(c) }
})

// PUT /admin/performer-pages/:slug/videos/:id
performerRoutes.put('/admin/performer-pages/:slug/videos/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    const body = await c.req.json<{
      title?: string; subtitle?: string; thumbnailUrl?: string; videoUrl?: string
    }>()
    const fields = {
      ...(body.title !== undefined && { title: body.title }),
      ...(body.subtitle !== undefined && { subtitle: body.subtitle }),
      ...(body.thumbnailUrl !== undefined && { thumbnailUrl: body.thumbnailUrl }),
      ...(body.videoUrl !== undefined && { videoUrl: body.videoUrl }),
    }
    const [row] = await db.update(performerVideos).set(fields).where(eq(performerVideos.id, id)).returning()
    if (!row) return notFound(c, 'video not found')
    return ok(c, row)
  } catch { return internalError(c) }
})

// DELETE /admin/performer-pages/:slug/videos/:id  (only non-main videos)
performerRoutes.delete('/admin/performer-pages/:slug/videos/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    const [video] = await db.select().from(performerVideos).where(eq(performerVideos.id, id)).limit(1)
    if (!video) return notFound(c, 'video not found')
    if (video.isMain) return badRequest(c, 'cannot delete main video')
    await db.delete(performerVideos).where(eq(performerVideos.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})
