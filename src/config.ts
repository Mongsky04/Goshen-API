// goshen-api/src/config.ts
import 'dotenv/config'

export const config = {
  port: process.env.PORT ?? '8080',
  databaseUrl: process.env.DATABASE_URL ?? '',
  jwtSecret: process.env.JWT_SECRET ?? 'change-me-in-production',
  frontendOrigin: process.env.FRONTEND_ORIGIN ?? 'http://localhost:3000',
  backendUrl: process.env.BACKEND_URL ?? 'http://localhost:8080',
  cloudinary: {
    cloudName: process.env.CLOUDINARY_CLOUD_NAME ?? '',
    apiKey: process.env.CLOUDINARY_API_KEY ?? '',
    apiSecret: process.env.CLOUDINARY_API_SECRET ?? '',
  },
} as const
