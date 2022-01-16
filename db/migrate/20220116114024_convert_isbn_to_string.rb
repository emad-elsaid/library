class ConvertIsbnToString < ActiveRecord::Migration[6.1]
  def up
    change_column :books, :isbn, :string, limit: 13
  end

  def down
    change_column :books, :isbn, :bigint
  end
end
