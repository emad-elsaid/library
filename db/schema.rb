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

ActiveRecord::Schema.define(version: 2021_05_20_122458) do

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
    t.integer "isbn"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.integer "shelf_id"
    t.integer "user_id"
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
    t.index ["slug"], name: "index_users_on_slug", unique: true
  end

  add_foreign_key "accesses", "users"
  add_foreign_key "accesses", "users", column: "owner_id"
  add_foreign_key "books", "shelves"
  add_foreign_key "books", "users"
  add_foreign_key "borrows", "books"
  add_foreign_key "borrows", "users"
  add_foreign_key "borrows", "users", column: "owner_id"
  add_foreign_key "shelves", "users"
end
