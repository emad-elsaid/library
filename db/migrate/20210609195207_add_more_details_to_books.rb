class AddMoreDetailsToBooks < ActiveRecord::Migration[6.1]
  def change
    add_column :books, :subtitle, :string
    add_column :books, :description, :string
    add_column :books, :page_count, :integer
    add_column :books, :publisher, :string
  end
end
