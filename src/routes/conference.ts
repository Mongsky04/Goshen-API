// goshen-api/src/routes/conference.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import {
  conferencePages, conferenceHero, conferenceSectionTitles,
  conferenceProducts, conferenceWorkspace, conferenceRoomSolutions,
  conferenceRoomKitItems, products,
} from '../db/schema.js'
import { ok, notFound, internalError, badRequest } from '../lib/response.js'
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
