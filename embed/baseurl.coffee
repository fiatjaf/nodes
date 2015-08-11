url = require 'url'

base = url.parse document.getElementById('nodes-embedded').src
base.pathname = ''
baseURL = url.format base
module.exports = baseURL
