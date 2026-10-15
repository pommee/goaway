# Building the GoAway documentation

The GoAway documentation uses [Material for MkDocs][1] and requires Python to
build.

```bash
python3 -m venv venv
. venv/bin/activate
pip install -r requirements.txt
make
```

To view the site in your local web browser, run

```bash
(sleep 2; python -m webbrowser -t http://localhost:8000) &
make serve
```

[1]: https://github.com/squidfunk/mkdocs-material
