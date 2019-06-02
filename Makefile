clean:
	@rm -rf example
	@go clean

# Tangle out the example in the README.org using Emacs org-babel
example:
	@docker run -it --rm -v $$(pwd):/haunter silex/emacs bash -c "cd /haunter ; install -o `stat -c '%u' ./README.org` -g `stat -c '%g' ./README.org` -d -m 755 /haunter/example ; emacs --batch -l org --eval '(org-babel-tangle-file \"/haunter/README.org\")'"
	@echo "See the example implementation in the 'example' directory!"

test:
	@go test -v
