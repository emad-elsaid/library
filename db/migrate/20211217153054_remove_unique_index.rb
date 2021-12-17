class RemoveUniqueIndex < ActiveRecord::Migration[6.1]
  def change
    remove_index :emails, [:emailable_type, :emailable_id], unique: true
    add_index :emails, [:emailable_type, :emailable_id]
  end
end
