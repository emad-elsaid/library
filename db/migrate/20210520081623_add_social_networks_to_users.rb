class AddSocialNetworksToUsers < ActiveRecord::Migration[6.1]
  def change
    profiles = [:facebook, :twitter, :linkedin, :instagram, :phone, :whatsapp, :telegram]
    profiles.each do |profile|
      add_column :users, profile, :string
    end
  end
end
