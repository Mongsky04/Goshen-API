// goshen-api/src/lib/response.ts
import type { Context } from 'hono'

export const ok = (c: Context, data: unknown) =>
  c.json({ success: true, data }, 200)

export const created = (c: Context, data: unknown) =>
  c.json({ success: true, data }, 201)

export const badRequest = (c: Context, error: string) =>
  c.json({ success: false, error }, 400)

export const unauthorized = (c: Context, error: string) =>
  c.json({ success: false, error }, 401)

export const notFound = (c: Context, error: string) =>
  c.json({ success: false, error }, 404)

export const internalError = (c: Context) =>
  c.json({ success: false, error: 'internal server error' }, 500)
