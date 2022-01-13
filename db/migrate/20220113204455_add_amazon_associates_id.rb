class AddAmazonAssociatesId < ActiveRecord::Migration[6.1]
  def change
    add_column :users, :amazon_associates_id, :string
  end
end
