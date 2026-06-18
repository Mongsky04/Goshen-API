// goshen-api/src/routes/conference.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import {
  conferencePages, conferenceHero, conferenceSectionTitles,
  conferenceProducts, conferenceWorkspace, conferenceRoomSolutions,
  conferenceRoomKitItems, products,
} from '../db/schema.js'
import { ok, created, notFound, internalError, badRequest } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'

const VALID_SLUGS = new Set(['enterprise', 'government', 'higher-education', 'hospitality'])

interface ConferencePutBody {
  isPublished?: boolean
  hero?: { heroImageUrl?: string; badgeText?: string; headline?: string; subText?: string }
  titles?: { productGrid?: string; workspace?: string; solutions?: string; contact?: string }
  workspaceDescription?: string
}

// Builds a full conference page object from DB
async function buildConferencePage(slug: string) {
  const [page] = await db.select().from(conferencePages).where(eq(conferencePages.slug, slug)).limit(1)
  if (!page) return null

  const [hero] = await db.select().from(conferenceHero).where(eq(conferenceHero.pageId, page.id)).limit(1)
  const titleRows = await db.select().from(conferenceSectionTitles)
    .where(eq(conferenceSectionTitles.pageId, page.id))
  const [workspace] = await db.select().from(conferenceWorkspace)
    .where(eq(conferenceWorkspace.pageId, page.id)).limit(1)

  const solutionRows = await db.select().from(conferenceRoomSolutions)
    .where(eq(conferenceRoomSolutions.pageId, page.id))
    .orderBy(asc(conferenceRoomSolutions.sortOrder))

  const solutionsWithKits = await Promise.all(solutionRows.map(async (s) => {
    const items = await db.select().from(conferenceRoomKitItems)
      .where(eq(conferenceRoomKitItems.roomSolutionId, s.id))
      .orderBy(asc(conferenceRoomKitItems.sortOrder))
    return { ...s, kitItems: items.map(i => i.item) }
  }))

  const productRows = await db
    .select({
      id: conferenceProducts.id,
      productId: conferenceProducts.productId,
      section: conferenceProducts.section,
      isHidden: conferenceProducts.isHidden,
      sortOrder: conferenceProducts.sortOrder,
      name: products.name,
      category: products.category,
      subCategory: products.subCategory,
      imageUrl: products.imageUrl,
    })
    .from(conferenceProducts)
    .innerJoin(products, eq(conferenceProducts.productId, products.id))
    .where(eq(conferenceProducts.pageId, page.id))
    .orderBy(asc(conferenceProducts.sortOrder))

  const titlesMap: Record<string, string> = {}
  for (const t of titleRows) titlesMap[t.sectionKey] = t.title

  return {
    id: page.id,
    slug: page.slug,
    label: page.label,
    isPublished: page.isPublished,
    hero: hero ? {
      heroImageUrl: hero.heroImageUrl,
      badgeText: hero.badgeText,
      headline: hero.headline,
      subText: hero.subText,
    } : null,
    titles: {
      productGrid: titlesMap['product_grid'] ?? '',
      workspace: titlesMap['workspace'] ?? '',
      solutions: titlesMap['solutions'] ?? '',
      contact: titlesMap['contact'] ?? '',
    },
    workspaceDescription: workspace?.description ?? '',
    solutions: solutionsWithKits,
    products: productRows,
  }
}

export const conferenceRoutes = new Hono()

// GET /api/v1/conference-pages/:slug (public)
conferenceRoutes.get('/conference-pages/:slug', async (c) => {
  try {
    const data = await buildConferencePage(c.req.param('slug'))
    if (!data) return notFound(c, 'conference page not found')
    return ok(c, data)
  } catch { return internalError(c) }
})

// GET /api/v1/admin/conference-pages/:slug (protected)
conferenceRoutes.get('/admin/conference-pages/:slug', requireAuth, async (c) => {
  try {
    const data = await buildConferencePage(c.req.param('slug'))
    if (!data) return notFound(c, 'conference page not found')
    return ok(c, data)
  } catch { return internalError(c) }
})

// PUT /api/v1/admin/conference-pages/:slug (protected)
// Accepts JSON body: { isPublished, hero, titles, workspaceDescription }
conferenceRoutes.put('/admin/conference-pages/:slug', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    if (!VALID_SLUGS.has(slug)) return badRequest(c, 'invalid conference slug')

    const body = await c.req.json<ConferencePutBody>()

    const [page] = await db.select().from(conferencePages).where(eq(conferencePages.slug, slug)).limit(1)
    if (!page) return notFound(c, 'conference page not found')

    const hero = body.hero ?? {}
    await db.insert(conferenceHero).values({
      pageId: page.id,
      heroImageUrl: hero.heroImageUrl ?? '',
      badgeText: hero.badgeText ?? '',
      headline: hero.headline ?? '',
      subText: hero.subText ?? '',
    }).onConflictDoUpdate({
      target: conferenceHero.pageId,
      set: {
        heroImageUrl: hero.heroImageUrl ?? '',
        badgeText: hero.badgeText ?? '',
        headline: hero.headline ?? '',
        subText: hero.subText ?? '',
      },
    })

    // Section titles — map camelCase dashboard keys to snake_case section_key values
    const titles = body.titles ?? {}
    const titleEntries: [string, string][] = [
      ['product_grid', titles.productGrid ?? ''],
      ['workspace', titles.workspace ?? ''],
      ['solutions', titles.solutions ?? ''],
      ['contact', titles.contact ?? ''],
    ]
    for (const [sectionKey, title] of titleEntries) {
      await db.insert(conferenceSectionTitles)
        .values({ pageId: page.id, sectionKey, title })
        .onConflictDoUpdate({
          target: [conferenceSectionTitles.pageId, conferenceSectionTitles.sectionKey],
          set: { title },
        })
    }

    const workspaceDesc = body.workspaceDescription ?? ''
    await db.insert(conferenceWorkspace).values({ pageId: page.id, description: workspaceDesc })
      .onConflictDoUpdate({ target: conferenceWorkspace.pageId, set: { description: workspaceDesc } })

    if (body.isPublished !== undefined) {
      await db.update(conferencePages)
        .set({ isPublished: body.isPublished })
        .where(eq(conferencePages.id, page.id))
    }

    return ok(c, await buildConferencePage(slug))
  } catch {
    return internalError(c)
  }
})

// POST /admin/conference-pages/:slug/products
conferenceRoutes.post('/admin/conference-pages/:slug/products', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    const [page] = await db.select().from(conferencePages).where(eq(conferencePages.slug, slug)).limit(1)
    if (!page) return notFound(c, 'conference page not found')
    const { productId, section } = await c.req.json<{ productId: number; section: string }>()
    if (!productId || !section) return badRequest(c, 'productId and section required')
    const count = await db.select().from(conferenceProducts)
      .where(eq(conferenceProducts.pageId, page.id)).then(r => r.length)
    const [row] = await db.insert(conferenceProducts)
      .values({ pageId: page.id, productId, section, sortOrder: count })
      .returning()
    return created(c, row)
  } catch (e: any) {
    if (e?.code === '23505') return c.json({ success: false, error: 'product already on this page' }, 409)
    return internalError(c)
  }
})

// PATCH /admin/conference-pages/:slug/products/:id
conferenceRoutes.patch('/admin/conference-pages/:slug/products/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    const { isHidden } = await c.req.json<{ isHidden: boolean }>()
    const [row] = await db.update(conferenceProducts).set({ isHidden }).where(eq(conferenceProducts.id, id)).returning()
    if (!row) return notFound(c, 'not found')
    return ok(c, row)
  } catch { return internalError(c) }
})

// DELETE /admin/conference-pages/:slug/products/:id
conferenceRoutes.delete('/admin/conference-pages/:slug/products/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    await db.delete(conferenceProducts).where(eq(conferenceProducts.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})

// POST /admin/conference-pages/:slug/solutions
conferenceRoutes.post('/admin/conference-pages/:slug/solutions', requireAuth, async (c) => {
  try {
    const slug = c.req.param('slug')
    const [page] = await db.select().from(conferencePages).where(eq(conferencePages.slug, slug)).limit(1)
    if (!page) return notFound(c, 'conference page not found')
    const body = await c.req.json<{
      roomSize: string; title: string; description: string; kitLabel: string; kitItems: string[]
      imageUrl: string; imageUrl2: string
      card1Name: string; card1Category: string; card1SubCategory: string
      card2Name: string; card2Category: string; card2SubCategory: string
      isHidden: boolean
    }>()
    const count = await db.select().from(conferenceRoomSolutions)
      .where(eq(conferenceRoomSolutions.pageId, page.id)).then(r => r.length)
    const [sol] = await db.insert(conferenceRoomSolutions).values({
      pageId: page.id,
      roomSize: body.roomSize,
      title: body.title ?? '',
      description: body.description ?? '',
      kitLabel: body.kitLabel ?? 'IMX ROOM KIT 30:',
      imageUrl: body.imageUrl ?? '',
      imageUrl2: body.imageUrl2 ?? '',
      card1Name: body.card1Name ?? '',
      card1Category: body.card1Category ?? '',
      card1SubCategory: body.card1SubCategory ?? '',
      card2Name: body.card2Name ?? '',
      card2Category: body.card2Category ?? '',
      card2SubCategory: body.card2SubCategory ?? '',
      isHidden: body.isHidden ?? false,
      sortOrder: count,
    }).returning()
    const items = (body.kitItems ?? []).filter(Boolean)
    for (let i = 0; i < items.length; i++) {
      await db.insert(conferenceRoomKitItems).values({ roomSolutionId: sol.id, item: items[i], sortOrder: i })
    }
    return created(c, sol)
  } catch (e: any) {
    if (e?.code === '23505') return c.json({ success: false, error: 'roomSize already exists for this page' }, 409)
    return internalError(c)
  }
})

// PUT /admin/conference-pages/:slug/solutions/:id
conferenceRoutes.put('/admin/conference-pages/:slug/solutions/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    const body = await c.req.json<{
      roomSize?: string; title?: string; description?: string; kitLabel?: string; kitItems?: string[]
      imageUrl?: string; imageUrl2?: string
      card1Name?: string; card1Category?: string; card1SubCategory?: string
      card2Name?: string; card2Category?: string; card2SubCategory?: string
      isHidden?: boolean
    }>()
    const { kitItems } = body
    const fields = {
      ...(body.roomSize !== undefined && { roomSize: body.roomSize }),
      ...(body.title !== undefined && { title: body.title }),
      ...(body.description !== undefined && { description: body.description }),
      ...(body.kitLabel !== undefined && { kitLabel: body.kitLabel }),
      ...(body.imageUrl !== undefined && { imageUrl: body.imageUrl }),
      ...(body.imageUrl2 !== undefined && { imageUrl2: body.imageUrl2 }),
      ...(body.card1Name !== undefined && { card1Name: body.card1Name }),
      ...(body.card1Category !== undefined && { card1Category: body.card1Category }),
      ...(body.card1SubCategory !== undefined && { card1SubCategory: body.card1SubCategory }),
      ...(body.card2Name !== undefined && { card2Name: body.card2Name }),
      ...(body.card2Category !== undefined && { card2Category: body.card2Category }),
      ...(body.card2SubCategory !== undefined && { card2SubCategory: body.card2SubCategory }),
      ...(body.isHidden !== undefined && { isHidden: body.isHidden }),
    }
    const [sol] = await db.update(conferenceRoomSolutions).set(fields).where(eq(conferenceRoomSolutions.id, id)).returning()
    if (!sol) return notFound(c, 'solution not found')
    if (kitItems !== undefined) {
      await db.delete(conferenceRoomKitItems).where(eq(conferenceRoomKitItems.roomSolutionId, id))
      const items = kitItems.filter(Boolean)
      for (let i = 0; i < items.length; i++) {
        await db.insert(conferenceRoomKitItems).values({ roomSolutionId: id, item: items[i], sortOrder: i })
      }
    }
    return ok(c, sol)
  } catch { return internalError(c) }
})

// DELETE /admin/conference-pages/:slug/solutions/:id
conferenceRoutes.delete('/admin/conference-pages/:slug/solutions/:id', requireAuth, async (c) => {
  try {
    const id = c.req.param('id')
    await db.delete(conferenceRoomSolutions).where(eq(conferenceRoomSolutions.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})
