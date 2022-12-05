Rails.application.routes.draw do
  resources :products, only: [:index, :show]
  resources :categories, only: [:index]
end
