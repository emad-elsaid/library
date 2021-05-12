# frozen_string_literal: true

require 'rack/utils'

helpers do
  def h(text)
    Rack::Utils.escape_html(text)
  end
end
