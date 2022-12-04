class CategoriesController < ApplicationController
  def index
    respond_to do |format|
      format.json { render(json: categories ) }
    end
  end

  private

  def categories
    @categories ||= [
      { id: 1, title: "Shoes" },
      { id: 2, title: "Pants" }
    ]
  end
end
