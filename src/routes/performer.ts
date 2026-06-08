// goshen-api/src/routes/performer.ts
import { Hono } from 'hono'
import { and, eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { performerPages, performerProducts, performerVideos, products } from '../db/schema.js'
import { ok, notFound, internalError, badRequest } from '../lib/response.js'
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
