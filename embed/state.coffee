tl  = require 'talio'

baseURL = require './baseurl'

State = tl.StateFactory
  here:
    url: location.href
    title: document.querySelector('title').innerHTML
  otherTabs: {}

module.exports = State
