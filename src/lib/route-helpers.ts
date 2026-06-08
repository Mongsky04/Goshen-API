// goshen-api/src/lib/route-helpers.ts
import { uploadFile, allowedMime } from './storage.js'

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10 MB

export class MimeError extends Error {
  readonly code = 'BAD_MIME'
  constructor(msg = 'unsupported file type') { super(msg) }
}

export function parsePage(query: (k: string) => string | undefined) {
  const page = Math.max(1, parseInt(query('page') ?? '1'))
  const limit = Math.min(100, Math.max(1, parseInt(query('limit') ?? '20')))
  return { page, limit, offset: (page - 1) * limit }
}

export async function uploadIfPresent(form: FormData, field: string): Promise<string | null> {
  const file = form.get(field) as File | null
  if (!file || file.size === 0) return null
  if (!allowedMime.has(file.type)) throw new MimeError()
  if (file.size > MAX_FILE_SIZE) throw new MimeError('file too large (max 10 MB)')
  return uploadFile(file)
}
