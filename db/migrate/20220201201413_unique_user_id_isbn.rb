class UniqueUserIdIsbn < ActiveRecord::Migration[6.1]
  def change
    add_index :books, [:user_id, :isbn], unique: true
  end
end
