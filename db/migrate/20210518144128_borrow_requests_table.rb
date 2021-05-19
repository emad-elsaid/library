class BorrowRequestsTable < ActiveRecord::Migration[6.1]
  def change
    create_table :borrows do |t|
      t.belongs_to :user, foreign_key: true, null: false
      t.belongs_to :book, foreign_key: true, null: false
      t.belongs_to :owner, foreign_key: { to_table: :users }, null: false
      t.integer :days
      t.timestamp :borrowed_at
      t.timestamp :returned_at

      t.timestamps
    end
  end
end
