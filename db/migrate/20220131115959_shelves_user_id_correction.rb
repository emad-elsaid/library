class ShelvesUserIdCorrection < ActiveRecord::Migration[6.1]
  def change
    change_column_null :shelves, :user_id, false
    change_column :shelves, :user_id, :bigint
  end
end
