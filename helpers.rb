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
      when :borrow then loggedin? && record.user != current_user && !record.borrows.exists?(user: current_user)
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Shelf
      case verb
      when :create then record.user == current_user
      when :edit then record.user == current_user
      when :delete then record.user == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    when User
      case verb
      when :edit then record == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Borrow
      case verb
      when :show then loggedin?
      when :delete then (record.user == current_user && record.borrowed_at.nil?) || record.owner == current_user
      when :borrow then !record.book.borrows.borrowed.exists? && record.owner == current_user
      when :return then record.borrowed_at.present? && record.owner == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    else
      raise "Error #{record} permissions not handled."
    end
  end

  def format_date(date)
    date.to_date
  end
end
