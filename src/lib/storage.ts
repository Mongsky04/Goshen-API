// goshen-api/src/lib/storage.ts
import { v2 as cloudinary } from 'cloudinary'
import { config } from '../config.js'

cloudinary.config({
  cloud_name: config.cloudinary.cloudName,
  api_key: config.cloudinary.apiKey,
  api_secret: config.cloudinary.apiSecret,
})

export const allowedMime = new Set([
  'image/jpeg', 'image/png', 'image/webp', 'image/gif', 'image/svg+xml',
])

export async function ensureCloudinaryFolder(): Promise<void> {
  try {
    await cloudinary.api.create_folder('goshen')
  } catch {
    // folder already exists — ignore
  }
}

export async function uploadFile(file: File): Promise<string> {
  const buffer = Buffer.from(await file.arrayBuffer())
  return new Promise((resolve, reject) => {
    const stream = cloudinary.uploader.upload_stream(
      { folder: 'goshen', resource_type: 'auto' },
      (err, result) => {
        if (err || !result) return reject(err ?? new Error('cloudinary upload failed'))
        resolve(result.secure_url)
      },
    )
    stream.end(buffer)
  })
}

export async function deleteCloudinaryAsset(url: string): Promise<void> {
  if (!url.includes('res.cloudinary.com')) return
  const match = url.match(/\/upload\/(?:v\d+\/)?(.+)$/)
  if (!match) return
  const publicId = match[1].replace(/\.[^/.]+$/, '')
  await cloudinary.uploader.destroy(publicId)
}
