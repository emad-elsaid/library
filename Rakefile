require 'bundler'
Bundler.require(:default)

require 'sinatra/activerecord/rake'

ActiveRecord::Base.schema_format = :sql

namespace :db do
  task :load_config do
    set :database, { adapter: 'postgresql', encoding: 'unicode', pool: 5, url: ENV['DATABASE_URL'] }
  end
end
