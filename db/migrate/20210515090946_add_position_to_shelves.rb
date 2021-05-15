class AddPositionToShelves < ActiveRecord::Migration[6.1]
  def change
    add_column :shelves, :position, :integer
  end
end
