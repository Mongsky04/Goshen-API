// goshen-api/src/routes/auth.ts
import { Hono } from 'hono'
import { SignJWT } from 'jose'
import { compare } from 'bcryptjs'
import { eq } from 'drizzle-orm'
import { db } from '../db/client.js'
import { admins } from '../db/schema.js'
import { config } from '../config.js'
import { ok, badRequest, unauthorized, internalError } from '../lib/response.js'
import { requireAuth } from '../middleware/auth.js'

type Env = { Variables: { adminId: string | undefined } }

export const authRoutes = new Hono<Env>()

authRoutes.post('/admin/login', async (c) => {
  const body: { email?: string; password?: string } = await c.req.json().catch(() => ({}))
  if (!body.email || !body.password) return badRequest(c, 'email and password required')

  const DUMMY_HASH = '$2b$12$invalidhashpaddddddddddddddddddddddddddddddddddddddddd'

  let admin: typeof admins.$inferSelect | undefined
  try {
    ;[admin] = await db.select().from(admins).where(eq(admins.email, body.email)).limit(1)
  } catch {
    return internalError(c)
  }

  // Always run compare to prevent timing-based user enumeration
  const hash = admin?.passwordHash ?? DUMMY_HASH
  const valid = await compare(body.password, hash)
  if (!admin || !valid) return unauthorized(c, 'invalid credentials')

  const secret = new TextEncoder().encode(config.jwtSecret)
  const token = await new SignJWT({ sub: String(admin.id) })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('24h')
    .sign(secret)

  return ok(c, { token })
})

authRoutes.get('/admin/me', requireAuth, (c) => ok(c, { id: c.get('adminId') }))
