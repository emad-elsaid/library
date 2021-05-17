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

ActiveRecord::Schema.define(version: 2021_05_17_193731) do

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
    t.index ["slug"], name: "index_users_on_slug", unique: true
  end

  add_foreign_key "books", "shelves"
  add_foreign_key "books", "users"
  add_foreign_key "shelves", "users"
end
