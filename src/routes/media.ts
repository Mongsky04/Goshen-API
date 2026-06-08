// goshen-api/src/routes/media.ts
import { Hono } from 'hono'
import { eq, desc } from 'drizzle-orm'
import { db } from '../db/client.js'
import { mediaAssets } from '../db/schema.js'
import { ok, badRequest, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'
import { uploadIfPresent, MimeError } from '../lib/route-helpers.js'

export const mediaRoutes = new Hono()

// GET /api/v1/media (public)
mediaRoutes.get('/', async (c) => {
  try {
    const rows = await db.select().from(mediaAssets).orderBy(desc(mediaAssets.createdAt))
    return ok(c, rows)
  } catch { return internalError(c) }
})

// POST /api/v1/media (protected) — upload + save to library
mediaRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const file = form.get('file') as File | null
    if (!file || file.size === 0) return badRequest(c, 'no file provided')
    const url = await uploadIfPresent(form, 'file')
    if (!url) return badRequest(c, 'no file provided')
    const [row] = await db.insert(mediaAssets).values({
      filename: file.name,
      url,
      size: file.size,
      mimeType: file.type,
    }).returning()
    return ok(c, row)
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})

// DELETE /api/v1/media/:id (protected)
mediaRoutes.delete('/:id', requireAuth, async (c) => {
  try {
    const id = parseInt(c.req.param('id'))
    if (isNaN(id)) return badRequest(c, 'invalid id')
    await db.delete(mediaAssets).where(eq(mediaAssets.id, id))
    return ok(c, null)
  } catch { return internalError(c) }
})
