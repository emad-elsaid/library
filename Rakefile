require 'bundler'
Bundler.require(:default)

require 'sinatra/activerecord/rake'

ActiveRecord::Base.schema_format = :sql

namespace :db do
  task :load_config do
    load './main'
  end
end
