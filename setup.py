import setuptools

with open("requirements.txt", "r+", encoding="utf-8") as file:
    dependences = file.read().strip().split("\n")

with open("README.md", "r+", encoding="utf-8") as file:
    long_description = file.read()

with open("version", "r+", encoding="utf-8") as file:
    version = file.read().split("\n")[0].strip()


setuptools.setup(
    name="bedrock-chunk-diff",
    version=version,
    author="Minecraft Muti-Media Organization",
    author_email="TriM-Organization@hotmail.com",
    description="Delta update implement for Minecraft bedrock chunk.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/TriM-Organization/bedrock-chunk-diff",
    classifiers=[
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Go",
        "Operating System :: POSIX :: Linux",
        "Operating System :: Microsoft :: Windows",
        "Operating System :: MacOS",
    ],
    package_dir={"bedrockchunkdiff": "python/package"},
    packages=[
        i.replace("package", "bedrockchunkdiff", 1)
        for i in setuptools.find_packages(where="python")
    ],
    package_data={
        "bedrockchunkdiff": [
            "dynamic_libs/*.so",
            "dynamic_libs/*.dll",
            "dynamic_libs/*.dylib",
        ],
    },
    install_requires=dependences,
    python_requires=">=3.10",
)
