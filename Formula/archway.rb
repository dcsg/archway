class Archway < Formula
  desc "Terraform for Code Architecture"
  homepage "https://github.com/dcsg/archway"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dcsg/archway/releases/download/v#{version}/archway_#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA"
    else
      url "https://github.com/dcsg/archway/releases/download/v#{version}/archway_#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/dcsg/archway/releases/download/v#{version}/archway_#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA"
    else
      url "https://github.com/dcsg/archway/releases/download/v#{version}/archway_#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA"
    end
  end

  def install
    bin.install "archway"
  end

  test do
    assert_match "archway version", shell_output("#{bin}/archway version")
  end
end
