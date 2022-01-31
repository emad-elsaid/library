class BooksUserIdToBigint < ActiveRecord::Migration[6.1]
  def up
    change_column :books, :user_id, :bigint
  end
  def down
    change_column :books, :user_id, :integer
  end
end
