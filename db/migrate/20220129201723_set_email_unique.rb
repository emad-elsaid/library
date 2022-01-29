class SetEmailUnique < ActiveRecord::Migration[6.1]
  def change
    add_index :users, :email, unique: true
  end
end
