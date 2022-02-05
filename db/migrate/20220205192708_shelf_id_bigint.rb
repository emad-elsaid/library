class ShelfIdBigint < ActiveRecord::Migration[6.1]
  def up
    change_column :books, :shelf_id, :bigint
  end
  def down
    change_column :books, :shelf_id, :integer
  end
end
