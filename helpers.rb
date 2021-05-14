# frozen_string_literal: true

require 'rack/utils'

helpers do
  def h(text)
    Rack::Utils.escape_html(text)
  end

  def current_user
    return unless session[:user]

    @current_user ||= User.find(session[:user])
  end
end
