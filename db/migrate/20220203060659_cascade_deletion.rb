class CascadeDeletion < ActiveRecord::Migration[6.1]
  def change
    remove_foreign_key :highlights, :books
    add_foreign_key :highlights, :books, on_delete: :cascade

    remove_foreign_key :books, :shelves
    add_foreign_key :books, :shelves, on_delete: :cascade

    remove_foreign_key :shelves, :users
    add_foreign_key :shelves, :users, on_delete: :cascade

    remove_foreign_key :books, :users
    add_foreign_key :books, :users, on_delete: :cascade
  end
end
