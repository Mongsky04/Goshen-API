// goshen-api/src/routes/articles.ts
import { Hono } from 'hono'
import { eq, asc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { articles } from '../db/schema.js'
import { ok, created, badRequest, notFound, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { parsePage, uploadIfPresent, MimeError } from '../lib/route-helpers.js'
import { deleteCloudinaryAsset } from '../lib/storage.js'

export const articleRoutes = new Hono()

articleRoutes.get('/', async (c) => {
  try {
    const { page, limit, offset } = parsePage(c.req.query.bind(c.req))
    const rows = await db.select().from(articles)
      .orderBy(asc(articles.sortOrder), asc(articles.id))
      .limit(limit).offset(offset)
    return ok(c, { data: rows, page, limit, total: null })
  } catch {
    return internalError(c)
  }
})

articleRoutes.get('/:id', async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [row] = await db.select().from(articles).where(eq(articles.id, id)).limit(1)
    if (!row) return notFound(c, 'article not found')
    return ok(c, row)
  } catch {
    return internalError(c)
  }
})

articleRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const title = form.get('title') as string
    if (!title) return badRequest(c, 'title is required')
    const description = (form.get('description') as string) ?? ''
    const rawDate = form.get('published_at') as string | null
    const publishedAt = rawDate ? new Date(rawDate) : new Date()
    const uploadedImageUrl = await uploadIfPresent(form, 'image')
    const directImageUrl = (form.get('image_url') as string) ?? ''
    const imageUrl = uploadedImageUrl ?? directImageUrl ?? ''
    const [row] = await db.insert(articles).values({ title, description, imageUrl, publishedAt }).returning()
    return created(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

articleRoutes.put('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [existing] = await db.select().from(articles).where(eq(articles.id, id)).limit(1)
    if (!existing) return notFound(c, 'article not found')
    const form = await c.req.formData()
    const title = (form.get('title') as string) || existing.title
    const description = (form.get('description') as string) ?? existing.description
    const rawDate = form.get('published_at') as string | null
    const publishedAt = rawDate ? new Date(rawDate) : existing.publishedAt
    const uploadedImageUrl = await uploadIfPresent(form, 'image')
    const directImageUrl = (form.get('image_url') as string) ?? ''
    const imageUrl = uploadedImageUrl ?? directImageUrl ?? existing.imageUrl
    const [row] = await db.update(articles).set({ title, description, imageUrl, publishedAt }).where(eq(articles.id, id)).returning()
    return ok(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

articleRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    const [deleted] = await db.delete(articles).where(eq(articles.id, id)).returning()
    if (deleted?.imageUrl) await deleteCloudinaryAsset(deleted.imageUrl)
    return ok(c, null)
  } catch {
    return internalError(c)
  }
})
