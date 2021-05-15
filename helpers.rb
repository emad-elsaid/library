# frozen_string_literal: true

require 'rack/utils'

helpers do
  def h(text)
    Rack::Utils.escape_html(text)
  end

  def guest?
    current_user.nil?
  end

  def loggedin?
    !current_user.nil?
  end

  def current_user
    return unless session[:user]

    @current_user ||= User.find(session[:user])
  end

  def can?(verb, record = nil)
    case record

    when nil
      case verb
      when :login then guest?
      when :logout then loggedin?
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Book
      case verb
      when :create then record.user == current_user
      when :edit then record.user == current_user
      when :delete then record.user == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Shelf
      case verb
      when :create then record.user == current_user
      when :edit then record.user == current_user
      when :delete then record.user == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    else
      raise "Error #{record} permissions not handled."
    end
  end
end
