class DropEmailsTable < ActiveRecord::Migration[6.1]
  def change
    drop_table "emails", force: :cascade do |t|
      t.bigint "user_id", null: false
      t.string "emailable_type"
      t.bigint "emailable_id"
      t.string "about", null: false
      t.datetime "created_at", precision: 6, null: false
      t.datetime "updated_at", precision: 6, null: false
      t.index ["emailable_type", "emailable_id"], name: "index_emails_on_emailable_type_and_emailable_id"
      t.index ["user_id"], name: "index_emails_on_user_id"
    end
  end
end
