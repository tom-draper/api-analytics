class ProductsController < ApplicationController
  def index
    respond_to do |format|
      format.json { render(json: products ) }
    end
  end

  def show
    respond_to do |format|
      format.json { render(json: {} ) }
    end
  end

  private

  def products
    @products ||= [
      { id: 1, title: "Soccer Ball", sku: "1234567" },
      { id: 2, title: "Volley Ball", sku: "2345678" },
      { id: 3, title: "Basket Ball", sku: "3456789" },
    ]
  end
end
