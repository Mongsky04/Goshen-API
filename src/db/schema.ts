// goshen-api/src/db/schema.ts
import { sql } from 'drizzle-orm'
import {
  pgTable, bigserial, text, integer, boolean, timestamp,
  bigint, uuid, unique, check,
} from 'drizzle-orm/pg-core'

const timestamps = {
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow(),
}

export const admins = pgTable('admins', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  email: text('email').notNull().unique(),
  passwordHash: text('password_hash').notNull(),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
})

export const slider = pgTable('slider', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  title: text('title').notNull().default(''),
  imageUrl: text('image_url').notNull().default(''),
  orderNum: integer('order_num').notNull().default(0),
  ...timestamps,
})

export const products = pgTable('products', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  name: text('name').notNull(),
  imageUrl: text('image_url').notNull().default(''),
  category: text('category').notNull().default(''),
  subCategory: text('sub_category').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const featured = pgTable('featured', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  productId: bigint('product_id', { mode: 'number' }).notNull().references(() => products.id, { onDelete: 'cascade' }),
  name: text('name').notNull().default(''),
  imageUrl: text('image_url').notNull().default(''),
  category: text('category').notNull().default(''),
  subCategory: text('sub_category').notNull().default(''),
  featuredCategories: text('featured_categories').array().notNull().default([]),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const homepageGridProducts = pgTable('homepage_grid_products', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  productId: bigint('product_id', { mode: 'number' }).notNull().references(() => products.id, { onDelete: 'cascade' }),
  sortOrder: integer('sort_order').notNull().default(0),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
})

export const banners = pgTable('banners', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  name: text('name').notNull().default(''),
  imageUrl: text('image_url').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const articles = pgTable('articles', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  title: text('title').notNull(),
  description: text('description').notNull().default(''),
  imageUrl: text('image_url').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  publishedAt: timestamp('published_at', { withTimezone: true }).notNull().defaultNow(),
  ...timestamps,
})

export const brands = pgTable('brands', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  name: text('name').notNull(),
  imageUrl: text('image_url').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const customers = pgTable('customers', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  imageUrl: text('image_url').notNull().default(''),
  altText: text('alt_text').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const homepageSupportCards = pgTable('homepage_support_cards', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: text('title').notNull().default(''),
  description: text('description').notNull().default(''),
  ctaLabel: text('cta_label').notNull().default(''),
  ctaHref: text('cta_href').notNull().default('#'),
  sortOrder: integer('sort_order').notNull().default(0),
})

export const conferencePages = pgTable('conference_pages', {
  id: uuid('id').primaryKey().defaultRandom(),
  slug: text('slug').notNull().unique(),
  label: text('label').notNull(),
  isPublished: boolean('is_published').notNull().default(false),
  ...timestamps,
})

export const conferenceHero = pgTable('conference_hero', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => conferencePages.id, { onDelete: 'cascade' }),
  heroImageUrl: text('hero_image_url').notNull().default(''),
  badgeText: text('badge_text').notNull().default(''),
  headline: text('headline').notNull().default(''),
  subText: text('sub_text').notNull().default(''),
  ...timestamps,
}, (t) => [unique().on(t.pageId)])

export const conferenceSectionTitles = pgTable('conference_section_titles', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => conferencePages.id, { onDelete: 'cascade' }),
  sectionKey: text('section_key').notNull(),
  title: text('title').notNull().default(''),
  ...timestamps,
}, (t) => [unique().on(t.pageId, t.sectionKey)])

export const conferenceProducts = pgTable('conference_products', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => conferencePages.id, { onDelete: 'cascade' }),
  section: text('section').notNull(),
  productId: bigint('product_id', { mode: 'number' }).notNull().references(() => products.id, { onDelete: 'cascade' }),
  isHidden: boolean('is_hidden').notNull().default(false),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
}, (t) => [
  check('section_check', sql`${t.section} IN ('product_grid', 'workspace', 'solutions')`),
])

export const conferenceWorkspace = pgTable('conference_workspace', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => conferencePages.id, { onDelete: 'cascade' }),
  description: text('description').notNull().default(''),
  ...timestamps,
}, (t) => [unique().on(t.pageId)])

export const conferenceRoomSolutions = pgTable('conference_room_solutions', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => conferencePages.id, { onDelete: 'cascade' }),
  roomSize: text('room_size').notNull(),
  title: text('title').notNull().default(''),
  description: text('description').notNull().default(''),
  kitLabel: text('kit_label').notNull().default('IMX ROOM KIT 30:'),
  imageUrl: text('image_url').notNull().default(''),
  imageUrl2: text('image_url_2').notNull().default(''),
  card1Name: text('card1_name').notNull().default(''),
  card1Category: text('card1_category').notNull().default(''),
  card1SubCategory: text('card1_sub_category').notNull().default(''),
  card2Name: text('card2_name').notNull().default(''),
  card2Category: text('card2_category').notNull().default(''),
  card2SubCategory: text('card2_sub_category').notNull().default(''),
  isHidden: boolean('is_hidden').notNull().default(false),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
}, (t) => [unique().on(t.pageId, t.roomSize)])

export const conferenceRoomKitItems = pgTable('conference_room_kit_items', {
  id: uuid('id').primaryKey().defaultRandom(),
  roomSolutionId: uuid('room_solution_id').notNull().references(() => conferenceRoomSolutions.id, { onDelete: 'cascade' }),
  item: text('item').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const performerPages = pgTable('performer_pages', {
  id: uuid('id').primaryKey().defaultRandom(),
  slug: text('slug').notNull().unique(),
  label: text('label').notNull(),
  isPublished: boolean('is_published').notNull().default(false),
  heroImageUrl: text('hero_image_url').notNull().default(''),
  productGridTitle: text('product_grid_title').notNull().default(''),
  videosSectionTitle: text('videos_section_title').notNull().default(''),
  ...timestamps,
})

export const performerProducts = pgTable('performer_products', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => performerPages.id, { onDelete: 'cascade' }),
  productId: bigint('product_id', { mode: 'number' }).notNull().references(() => products.id, { onDelete: 'cascade' }),
  isHidden: boolean('is_hidden').notNull().default(false),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const performerVideos = pgTable('performer_videos', {
  id: uuid('id').primaryKey().defaultRandom(),
  pageId: uuid('page_id').notNull().references(() => performerPages.id, { onDelete: 'cascade' }),
  isMain: boolean('is_main').notNull().default(false),
  title: text('title').notNull().default(''),
  subtitle: text('subtitle').notNull().default(''),
  thumbnailUrl: text('thumbnail_url').notNull().default(''),
  videoUrl: text('video_url').notNull().default(''),
  sortOrder: integer('sort_order').notNull().default(0),
  ...timestamps,
})

export const mediaAssets = pgTable('media_assets', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  filename: text('filename').notNull().default(''),
  url: text('url').notNull(),
  size: bigint('size', { mode: 'number' }).notNull().default(0),
  mimeType: text('mime_type').notNull().default(''),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
})

export const pageBanners = pgTable('page_banners', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  pageSlug: text('page_slug').notNull(),
  bannerId: bigint('banner_id', { mode: 'number' }).notNull().references(() => slider.id, { onDelete: 'cascade' }),
  orderNum: integer('order_num').notNull().default(0),
}, (t) => [unique().on(t.pageSlug, t.bannerId)])
