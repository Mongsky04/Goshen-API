// goshen-api/src/lib/storage.ts
import { v2 as cloudinary } from 'cloudinary'
import { createWriteStream, mkdirSync } from 'node:fs'
import { join, extname } from 'node:path'
import { randomUUID } from 'node:crypto'
import { config } from '../config.js'

const hasCloudinary = !!(config.cloudinary.cloudName && config.cloudinary.apiKey && config.cloudinary.apiSecret)

if (hasCloudinary) {
  cloudinary.config({
    cloud_name: config.cloudinary.cloudName,
    api_key: config.cloudinary.apiKey,
    api_secret: config.cloudinary.apiSecret,
  })
}

mkdirSync(config.uploadDir, { recursive: true })

export const allowedMime = new Set([
  'image/jpeg', 'image/png', 'image/webp', 'image/gif', 'image/svg+xml',
])

export async function uploadFile(file: File): Promise<string> {
  const buffer = Buffer.from(await file.arrayBuffer())

  if (hasCloudinary) {
    return new Promise((resolve, reject) => {
      const stream = cloudinary.uploader.upload_stream(
        { resource_type: 'auto' },
        (err, result) => {
          if (err || !result) return reject(err ?? new Error('cloudinary upload failed'))
          resolve(result.secure_url)
        },
      )
      stream.end(buffer)
    })
  }

  // Local fallback
  const ext = extname(file.name) || '.bin'
  const filename = `${randomUUID()}${ext}`
  const dest = join(config.uploadDir, filename)
  await new Promise<void>((resolve, reject) => {
    const ws = createWriteStream(dest)
    ws.write(buffer)
    ws.end()
    ws.on('finish', resolve)
    ws.on('error', reject)
  })
  return `${config.backendUrl}/uploads/${filename}`
}
