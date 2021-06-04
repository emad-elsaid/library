class CreateAccessTable < ActiveRecord::Migration[6.1]
  def change
    create_table :accesses do |t|
      t.references :user, foreign_key: true, null: false
      t.references :owner, null: false
      t.timestamp :accepted_at, index: true
      t.timestamp :rejected_at, index: true
      t.foreign_key :users, column: :owner_id

      t.timestamps
    end
  end
end
