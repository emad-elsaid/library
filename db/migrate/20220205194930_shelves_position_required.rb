class ShelvesPositionRequired < ActiveRecord::Migration[6.1]
  def change
    change_column_null :shelves, :position,  false
  end
end
