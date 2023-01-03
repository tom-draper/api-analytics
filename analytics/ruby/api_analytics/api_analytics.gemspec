# frozen_string_literal: true

require_relative "lib/api_analytics/version"

Gem::Specification.new do |spec|
  spec.name          = "api_analytics"
  spec.version       = Analytics::VERSION
  spec.authors       = ["Tom Draper"]
  spec.email         = ["tomjdraper1@gmail.com"]
  spec.summary       = "Monitoring and analytics for API applications."
  spec.description   = "Monitoring and analytics for API applications."
  spec.homepage      = "https://github.com/tom-draper/api-analytics"
  spec.license       = "MIT"
  spec.required_ruby_version = ">= 2.4.0"

  spec.metadata["homepage_uri"] = spec.homepage
  spec.metadata["source_code_uri"] = "https://github.com/tom-draper/api-analytics"

  # Specify which files should be added to the gem when it is released.
  # The `git ls-files -z` loads the files in the RubyGem that have been added into git.
  spec.files = Dir.chdir(File.expand_path(__dir__)) do
    `git ls-files -z`.split("\x0").reject { |f| f.match(%r{\A(?:test|spec|features)/}) }
  end
  spec.bindir        = "exe"
  spec.executables   = spec.files.grep(%r{\Aexe/}) { |f| File.basename(f) }
  spec.require_paths = ["lib"]
end
