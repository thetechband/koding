class EnvironmentContainer extends KDDiaContainer

  constructor:(options={}, data)->

    options.cssClass   = 'environments-container'
    # options.draggable  = yes

    super options, data

    title   = @getOption 'title'
    @header = new KDHeaderView {type : "medium", title}

    @itemHeight = options.itemHeight ? 40

    @on "DataLoaded", => @_dataLoaded = yes
    # @on "DragFinished", @bound 'savePosition'

    @newItemPlus = new KDCustomHTMLView
      cssClass   : 'new-item-plus'
      partial    : "<i></i><span>Add new</span>"
      click      : =>
        @once 'transitionend', @emit 'PlusButtonClicked'

    @newItemPlus.bindTransitionEnd()

    @loader = new KDLoaderView
      cssClass   : 'new-item-loader hidden'
      size       :
        height   : 20
        width    : 20

  viewAppended:->
    super

    @addSubView @header
    @header.addSubView @newItemPlus
    @header.addSubView @loader

    {@appStorage} = @parent
    # @appStorage.ready @bound 'loadPosition'

  showLoader: ->
    @newItemPlus.hide()
    @loader.show()

  hideLoader: ->
    @newItemPlus.show()
    @loader.hide()

  addDia:(diaObj, pos)->
    pos = x: 20, y: 60 + @diaCount() * (@itemHeight + 10)
    super diaObj, pos

    diaObj.on "KDObjectWillBeDestroyed", @bound 'updatePositions'
    diaObj.on "KDObjectWillBeDestroyed", => @emit "itemRemoved"
    # @updateHeight()

  updatePositions:->

    index = 0
    for _key, dia of @dias
      dia.setX 20
      dia.setY 60 + index * 50
      index++
    # @updateHeight()

  diaCount:-> Object.keys(@dias).length

  # updateHeight:->

  #   @setHeight 80 + @diaCount() * 50
  #   @emit 'UpdateScene'

  # savePosition:->

  #   name      = @constructor.name
  #   bounds    = x: @getRelativeX(), y: @getRelativeY()
  #   positions = (@appStorage.getValue 'containerPositions') or {}
  #   positions[name] = bounds
  #   @appStorage.setValue 'containerPositions', positions

  # loadPosition:->

  #   name     = @constructor.name
  #   position = ((@appStorage.getValue 'containerPositions') or {})[name]
  #   return  unless position
  #   @setX position.x; @setY position.y

  # resetPosition:->

  #   @setX @_initialPosition.x
  #   @setY @_initialPosition.y

  #   name      = @constructor.name
  #   positions = (@appStorage.getValue 'containerPositions') or {}

  #   delete positions[name]
  #   @appStorage.setValue 'containerPositions', positions

  loadItems:->
    @removeAllItems()
