class RemoveAccessBorrows < ActiveRecord::Migration[6.1]
  def change
    drop_table "accesses", force: :cascade do |t|
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

    drop_table "borrows", force: :cascade do |t|
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
  end
end
