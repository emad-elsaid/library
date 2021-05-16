class AddSlugToUsers < ActiveRecord::Migration[6.1]
  def change
    add_column :users, :slug, :string, null: false
    add_index :users, :slug, unique: true
  end
end
