# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 2022_01_16_114024) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "accesses", force: :cascade do |t|
    t.integer "user_id", null: false
    t.integer "owner_id", null: false
    t.datetime "accepted_at"
    t.datetime "rejected_at"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.index ["accepted_at"], name: "index_accesses_on_accepted_at"
    t.index ["owner_id"], name: "index_accesses_on_owner_id"
    t.index ["rejected_at"], name: "index_accesses_on_rejected_at"
    t.index ["user_id"], name: "index_accesses_on_user_id"
  end

  create_table "books", force: :cascade do |t|
    t.string "title"
    t.string "author"
    t.string "image"
    t.string "isbn", limit: 13
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.integer "shelf_id"
    t.integer "user_id"
    t.string "google_books_id"
    t.string "subtitle"
    t.string "description"
    t.integer "page_count"
    t.string "publisher"
    t.index ["shelf_id"], name: "index_books_on_shelf_id"
    t.index ["user_id"], name: "index_books_on_user_id"
  end

  create_table "borrows", force: :cascade do |t|
    t.integer "user_id", null: false
    t.integer "book_id", null: false
    t.integer "owner_id", null: false
    t.integer "days"
    t.datetime "borrowed_at"
    t.datetime "returned_at"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.index ["book_id"], name: "index_borrows_on_book_id"
    t.index ["owner_id"], name: "index_borrows_on_owner_id"
    t.index ["user_id"], name: "index_borrows_on_user_id"
  end

  create_table "emails", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "emailable_type"
    t.bigint "emailable_id"
    t.string "about", null: false
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.index ["emailable_type", "emailable_id"], name: "index_emails_on_emailable_type_and_emailable_id"
    t.index ["user_id"], name: "index_emails_on_user_id"
  end

  create_table "highlights", force: :cascade do |t|
    t.bigint "book_id", null: false
    t.integer "page", null: false
    t.string "content", null: false
    t.string "image"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.index ["book_id"], name: "index_highlights_on_book_id"
  end

  create_table "shelves", force: :cascade do |t|
    t.string "name"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.integer "user_id"
    t.integer "position"
    t.index ["user_id"], name: "index_shelves_on_user_id"
  end

  create_table "users", force: :cascade do |t|
    t.string "name"
    t.string "email"
    t.string "image"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.string "slug", null: false
    t.text "description"
    t.string "facebook"
    t.string "twitter"
    t.string "linkedin"
    t.string "instagram"
    t.string "phone"
    t.string "whatsapp"
    t.string "telegram"
    t.string "amazon_associates_id"
    t.index ["slug"], name: "index_users_on_slug", unique: true
  end

  add_foreign_key "accesses", "users"
  add_foreign_key "accesses", "users", column: "owner_id"
  add_foreign_key "books", "shelves"
  add_foreign_key "books", "users"
  add_foreign_key "borrows", "books"
  add_foreign_key "borrows", "users"
  add_foreign_key "borrows", "users", column: "owner_id"
  add_foreign_key "emails", "users"
  add_foreign_key "highlights", "books"
  add_foreign_key "shelves", "users"
end
