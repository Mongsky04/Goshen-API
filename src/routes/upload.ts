// goshen-api/src/routes/upload.ts
import { Hono } from 'hono'
import { requireAuth } from '../middleware/auth.js'
import { uploadIfPresent, MimeError } from '../lib/route-helpers.js'
import { badRequest, ok, internalError } from '../lib/response.js'

export const uploadRoutes = new Hono()

// POST /api/v1/upload — single image upload, returns { url }
uploadRoutes.post('/', requireAuth, async (c) => {
  try {
    const form = await c.req.formData()
    const url = await uploadIfPresent(form, 'file')
    if (!url) return badRequest(c, 'no file provided')
    return ok(c, { url })
  } catch (e) {
    if (e instanceof MimeError) return badRequest(c, e.message)
    return internalError(c)
  }
})
