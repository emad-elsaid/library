class AddForeignKeyShelfToBooks < ActiveRecord::Migration[6.1]
  def change
    add_foreign_key :books, :shelves
  end
end
