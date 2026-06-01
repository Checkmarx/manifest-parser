from setuptools import setup, find_packages

setup(
    name='my-package',
    version='0.1.0',
    install_requires=[
        'requests>=2.26.0',
        'flask==2.0.1',
        'six',
        'Pillow==9.0.0',
        'cryptography==2.9.2',
    ],
    extras_require={
        'dev': [
            'pytest>=6.0',
            'black==22.3.0',
        ],
    },
    tests_require=[
        'pytest',
    ],
)
