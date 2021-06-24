class CreateEmails < ActiveRecord::Migration[6.1]
  def change
    create_table :emails do |t|
      t.belongs_to :user, foreign_key: true, null: false
      t.belongs_to :emailable, polymorphic: true, index: { unique: true }

      t.timestamps
    end
  end
end
