// goshen-api/src/routes/case-study.ts
import { Hono } from 'hono'
import { asc, eq } from 'drizzle-orm'
import { db } from '../db/client.js'
import { caseStudies } from '../db/schema.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { parsePage } from '../lib/route-helpers.js'

export const caseStudyRoutes = new Hono()

// GET /api/v1/case-studies (public)
caseStudyRoutes.get('/', async (c) => {
  try {
    const { page, limit, offset } = parsePage(c.req.query.bind(c.req))
    const rows = await db.select().from(caseStudies)
      .orderBy(asc(caseStudies.sortOrder), asc(caseStudies.id))
      .limit(limit).offset(offset)
    return ok(c, { data: rows, page, limit, total: null })
  } catch { return internalError(c) }
})

// GET /api/v1/case-studies/:id (public)
caseStudyRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [row] = await db.select().from(caseStudies).where(eq(caseStudies.id, id)).limit(1)
    if (!row) return notFound(c, 'case study not found')
    return ok(c, row)
  } catch { return internalError(c) }
})

// POST /api/v1/case-studies (admin)
caseStudyRoutes.post('/', requireAuth, async (c) => {
  try {
    const body = await c.req.json<{ title?: string; description?: string; imageUrl?: string; isFeatured?: boolean }>()
    if (!body.title) return badRequest(c, 'title is required')
    const [row] = await db.insert(caseStudies).values({
      title: body.title,
      description: body.description ?? '',
      imageUrl: body.imageUrl ?? '',
      isFeatured: body.isFeatured ?? false,
    }).returning()
    return created(c, row)
  } catch { return internalError(c) }
})

// PUT /api/v1/case-studies/:id (admin)
caseStudyRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [existing] = await db.select().from(caseStudies).where(eq(caseStudies.id, id)).limit(1)
    if (!existing) return notFound(c, 'case study not found')
    const body = await c.req.json<{ title?: string; description?: string; imageUrl?: string; isFeatured?: boolean }>()
    const [row] = await db.update(caseStudies).set({
      title: body.title ?? existing.title,
      description: body.description ?? existing.description,
      imageUrl: body.imageUrl ?? existing.imageUrl,
      isFeatured: body.isFeatured ?? existing.isFeatured,
    }).where(eq(caseStudies.id, id)).returning()
    return ok(c, row)
  } catch { return internalError(c) }
})

// DELETE /api/v1/case-studies/:id (admin)
caseStudyRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    await db.delete(caseStudies).where(eq(caseStudies.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})
