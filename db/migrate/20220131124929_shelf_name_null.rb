class ShelfNameNull < ActiveRecord::Migration[6.1]
  def change
    change_column_null :shelves, :name, false
  end
end
