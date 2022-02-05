class ClearShelfOnDeletion < ActiveRecord::Migration[6.1]
  def change
    remove_foreign_key :books, :shelves
    add_foreign_key :books, :shelves, on_delete: :nullify
  end
end
