# frozen_string_literal: true

require 'rack/utils'

helpers do
  def partial (template, locals = {})
    erb(template.to_sym, layout: false, locals: locals)
  end

  def h(text)
    Rack::Utils.escape_html(text)
  end

  def simple_format(text)
    h(text).lines.join("<br/>")
  end

  def guest?
    current_user.nil?
  end

  def loggedin?
    !current_user.nil?
  end

  def book_cover(book)
    return "/books/image/#{book.image}" if book.image?
    return "https://books.google.com/books/content?id=#{book.google_books_id}&printsec=frontcover&img=1&zoom=1" if book.google_books_id?

    '/default_book'
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
      when :create then record.user_id == current_user&.id
      when :edit then record.user_id == current_user&.id
      when :delete then record.user_id == current_user&.id
      when :highlight then record.user_id == current_user&.id
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Shelf
      case verb
      when :create then record.user_id == current_user&.id
      when :edit then record.user_id == current_user&.id
      when :delete then record.user_id == current_user&.id
      else raise "Verb #{verb} not handled for #{record}"
      end

    when User
      case verb
      when :edit then record == current_user
      when :list_shelves then loggedin? && record == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Highlight
      case verb
      when :edit then record.book.user_id == current_user&.id
      when :delete then record.book.user_id == current_user&.id
      else raise "Verb #{verb} not handled for #{record}"
      end

    else raise "Error #{record} permissions not handled."
    end
  end

  def format_date(date)
    date.to_date
  end

  def meta_property(name)
    return unless @meta
    return unless @meta.key?(name)

    "<meta property=\"#{name}\" value=\"#{h @meta[name]}\"/>"
  end

  def meta_name(name)
    return unless @meta
    return unless @meta.key?(name)

    "<meta name=\"#{name}\" value=\"#{h @meta[name]}\"/>"
  end

  def set_meta(name, value)
    @meta ||= {}
    @meta[name] = value
  end
end
