// goshen-api/src/middleware/auth.ts
import { createMiddleware } from 'hono/factory'
import { jwtVerify } from 'jose'
import { config } from '../config.js'
import { unauthorized } from '../lib/response.js'

export const requireAuth = createMiddleware(async (c, next) => {
  const header = c.req.header('Authorization') ?? ''
  if (!header.startsWith('Bearer ')) return unauthorized(c, 'missing or invalid token')

  const token = header.slice(7)
  try {
    const secret = new TextEncoder().encode(config.jwtSecret)
    const { payload } = await jwtVerify(token, secret)
    if (!payload.sub) throw new Error('missing sub claim')
    c.set('adminId', payload.sub)
  } catch {
    return unauthorized(c, 'invalid token')
  }
  await next()
})
