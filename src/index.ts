// goshen-api/src/index.ts
import { serve } from '@hono/node-server'
import { serveStatic } from '@hono/node-server/serve-static'
import { Hono } from 'hono'
import { logger } from 'hono/logger'
import { config } from './config.js'
import { authRoutes } from './routes/auth.js'
import { productRoutes } from './routes/products.js'
import { featuredRoutes } from './routes/featured.js'
import { articleRoutes } from './routes/articles.js'
import { sliderRoutes } from './routes/sliders.js'
import { brandRoutes } from './routes/brands.js'
import { customerRoutes } from './routes/customers.js'
import { homepageRoutes } from './routes/homepage.js'
import { uploadRoutes } from './routes/upload.js'
import { mediaRoutes } from './routes/media.js'
import { conferenceRoutes } from './routes/conference.js'
import { performerRoutes } from './routes/performer.js'

const origins = config.frontendOrigin.split(',').map(s => s.trim()).filter(Boolean)

const app = new Hono()

app.use(logger())

// CORS — only emit header for explicitly allowed origins
app.use('*', async (c, next) => {
  const origin = c.req.header('Origin') ?? ''
  if (origins.includes(origin)) {
    c.header('Access-Control-Allow-Origin', origin)
    c.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS')
    c.header('Access-Control-Allow-Headers', 'Content-Type, Authorization')
  }
  if (c.req.method === 'OPTIONS') return c.body(null, 204)
  await next()
})

// Health check
app.get('/health', (c) => c.json({ status: 'ok' }))

// Serve local uploads
app.use('/uploads/*', serveStatic({ root: './' }))

// Auth routes (login, me)
app.route('/', authRoutes)

// Public + protected API routes
app.route('/api/v1/products', productRoutes)
app.route('/api/v1/featured', featuredRoutes)
app.route('/api/v1/articles', articleRoutes)
app.route('/api/v1/banners', sliderRoutes)
app.route('/api/v1/brands', brandRoutes)
app.route('/api/v1/customers', customerRoutes)
app.route('/api/v1', homepageRoutes)
app.route('/api/v1/upload', uploadRoutes)
app.route('/api/v1/media', mediaRoutes)
app.route('/api/v1', conferenceRoutes)
app.route('/api/v1', performerRoutes)

serve({ fetch: app.fetch, port: parseInt(config.port) }, (info) => {
  console.log(`Server started on port ${info.port}`)
})
