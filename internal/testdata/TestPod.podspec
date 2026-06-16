Pod::Spec.new do |s|
  s.name             = 'TestPod'
  s.version          = '1.0.0'
  s.summary          = 'A test pod fixture for the manifest parser.'
  s.homepage         = 'https://example.com/TestPod'
  s.license          = { :type => 'MIT', :file => 'LICENSE' }
  s.author           = { 'Test' => 'test@example.com' }
  s.source           = { :git => 'https://example.com/TestPod.git', :tag => '1.0.0' }
  s.ios.deployment_target = '13.0'
  s.source_files     = 'Sources/**/*.{swift,h,m}'

  # Pinned exact version
  s.dependency 'Alamofire', '~> 5.0'

  # Pinned exact version (double-quoted)
  s.dependency "SwiftyJSON", "4.3.0"

  # No version specifier -> latest
  s.dependency 'Quick'

  # Two version arguments (CocoaPods allows up to two)
  s.dependency 'Nimble', '>= 9.0', '< 10.0'
end
