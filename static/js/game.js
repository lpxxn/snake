// 游戏常量
const CELL_SIZE = 20
const GRID_COLOR = '#ddd'
const SNAKE_COLORS = [
  '#3498db', '#e74c3c', '#2ecc71', '#f39c12', '#9b59b6',
  '#1abc9c', '#d35400', '#c0392b', '#16a085', '#8e44ad'
]
const FOOD_COLOR = '#e74c3c'

// 游戏变量
let canvas, ctx
let socket
let playerId
let players = []
let foods = []
let messageTimeout

// 方向键代码
const Direction = {
  UP: 0,
  RIGHT: 1,
  DOWN: 2,
  LEFT: 3
}

// 初始化游戏
function init () {
  canvas = document.getElementById('gameCanvas')
  ctx = canvas.getContext('2d')

  // 生成随机玩家ID
  playerId = Date.now().toString() + Math.floor(Math.random() * 1000)

  // 连接WebSocket
  connectWebSocket()

  // 添加键盘事件监听
  window.addEventListener('keydown', handleKeyDown)

  // 初始化时绘制网格
  drawGrid()

  console.log('游戏初始化完成，玩家ID:', playerId)
}

// 连接WebSocket
function connectWebSocket () {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws?id=${playerId}`

  console.log('尝试连接WebSocket:', wsUrl)

  socket = new WebSocket(wsUrl)

  socket.onopen = function () {
    console.log('WebSocket连接已建立')
    // 请求玩家列表
    requestPlayerList()
    // 初始绘制游戏界面
    drawGame()
  }

  socket.onmessage = function (event) {
    console.log('收到WebSocket消息:', event.data)
    try {
      const message = JSON.parse(event.data)
      // 添加原始消息的调试输出
      console.log('解析后的消息对象:', JSON.stringify(message))
      handleMessage(message)
    } catch (error) {
      console.error('解析WebSocket消息时出错:', error)
    }
  }

  socket.onclose = function () {
    console.log('WebSocket连接已关闭')
    showMessage('与服务器的连接已断开，请刷新页面重试')
  }

  socket.onerror = function (error) {
    console.error('WebSocket错误:', error)
    showMessage('连接错误，请刷新页面重试')
  }
}

// 处理收到的消息
function handleMessage (message) {
  console.log('处理消息类型:', message.type)
  switch (message.type) {
    case 'player_join':
      // 修复：检查payload是否已经是对象
      const joinPayload = typeof message.payload === 'string' ? JSON.parse(message.payload) : message.payload
      console.log(`玩家加入: ${joinPayload.name} (${joinPayload.id})`)
      players.push(joinPayload)
      updatePlayerList()
      break

    case 'player_leave':
      // 修复：检查payload是否已经是对象
      const leavePayload = typeof message.payload === 'string' ? JSON.parse(message.payload) : message.payload
      console.log(`玩家离开: ${leavePayload.name} (${leavePayload.id})`)
      players = players.filter(p => p.id !== leavePayload.id)
      updatePlayerList()
      break

    case 'player_list':
      // 修复：检查payload是否已经是对象
      players = typeof message.payload === 'string' ? JSON.parse(message.payload) : message.payload
      updatePlayerList()
      break

    case 'game_state':
      try {
        // 修复：检查payload是否已经是对象
        const gameState = typeof message.payload === 'string' ? JSON.parse(message.payload) : message.payload
        console.log('收到游戏状态:', JSON.stringify(gameState))
        console.log('食物数量:', gameState.foods ? gameState.foods.length : 0)
        console.log('玩家数量:', gameState.players ? gameState.players.length : 0)

        // 添加更多调试信息
        if (gameState.foods && gameState.foods.length > 0) {
          console.log('第一个食物:', JSON.stringify(gameState.foods[0]))
        }
        if (gameState.players && gameState.players.length > 0) {
          console.log('第一个玩家:', JSON.stringify(gameState.players[0]))
        }

        players = gameState.players || []
        foods = gameState.foods || []
        drawGame()
      } catch (error) {
        console.error('处理游戏状态时出错:', error, message.payload)
      }
      break

    case 'player_dead':
      // 修复：检查payload是否已经是对象
      const deadPayload = typeof message.payload === 'string' ? JSON.parse(message.payload) : message.payload
      if (deadPayload.id === playerId) {
        showMessage('你的蛇撞到了障碍物！正在重生...')
      } else {
        showMessage(`玩家 ${deadPayload.name} 的蛇撞到了障碍物！`)
      }
      break
  }
}

// 处理键盘事件
function handleKeyDown (event) {
  // 处理空格键重新开始游戏
  if (event.key === ' ' || event.code === 'Space') {
    restartGame()
    return
  }

  let direction

  switch (event.key) {
    case 'ArrowUp':
    case 'w':
    case 'W':
      direction = Direction.UP
      break

    case 'ArrowRight':
    case 'd':
    case 'D':
      direction = Direction.RIGHT
      break

    case 'ArrowDown':
    case 's':
    case 'S':
      direction = Direction.DOWN
      break

    case 'ArrowLeft':
    case 'a':
    case 'A':
      direction = Direction.LEFT
      break

    default:
      return // 忽略其他按键
  }

  // 发送方向命令
  sendDirection(direction)
}

// 发送方向命令
function sendDirection (direction) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const message = {
      type: 'direction',
      payload: {
        direction: direction
      }
    }

    socket.send(JSON.stringify(message))
  }
}

// 重新开始游戏
function restartGame () {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const message = {
      type: 'restart',
      payload: {}
    }

    socket.send(JSON.stringify(message))
    showMessage('游戏重新开始')
  }
}

// 更新玩家列表
function updatePlayerList () {
  const playersList = document.getElementById('players')
  playersList.innerHTML = ''

  if (players.length === 0) {
    const li = document.createElement('li')
    li.textContent = '等待玩家加入...'
    playersList.appendChild(li)
    return
  }

  players.forEach((player, index) => {
    const li = document.createElement('li')
    const colorIndex = index % SNAKE_COLORS.length

    li.innerHTML = `
            <span style="color: ${SNAKE_COLORS[colorIndex]}">■</span>
            ${player.name} ${player.id === playerId ? '(你)' : ''}
        `

    playersList.appendChild(li)
  })
}

// 请求玩家列表
function requestPlayerList () {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const message = {
      type: 'get_players',
      payload: {}
    }

    socket.send(JSON.stringify(message))
  }
}

// 绘制游戏
function drawGame () {
  // 清空画布
  ctx.clearRect(0, 0, canvas.width, canvas.height)

  // 绘制网格
  drawGrid()

  // 绘制食物
  drawFoods()

  // 绘制所有蛇
  drawSnakes()
}

// 绘制网格
function drawGrid () {
  ctx.strokeStyle = GRID_COLOR
  ctx.lineWidth = 0.5

  // 绘制垂直线
  for (let x = 0; x <= canvas.width; x += CELL_SIZE) {
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, canvas.height)
    ctx.stroke()
  }

  // 绘制水平线
  for (let y = 0; y <= canvas.height; y += CELL_SIZE) {
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(canvas.width, y)
    ctx.stroke()
  }
}

// 绘制食物
function drawFoods () {
  ctx.fillStyle = FOOD_COLOR
  console.log('绘制食物数量:', foods.length)

  if (!foods || foods.length === 0) {
    console.log('没有食物可绘制')
    return
  }

  foods.forEach((food, index) => {
    if (!food || typeof food.x === 'undefined' || typeof food.y === 'undefined') {
      console.error('无效的食物对象:', food)
      return
    }

    console.log(`绘制食物 ${index}:`, food)
    // 确保坐标是数字
    const x = Number(food.x)
    const y = Number(food.y)

    if (isNaN(x) || isNaN(y)) {
      console.error('食物坐标无效:', food)
      return
    }

    ctx.fillRect(
      x * CELL_SIZE,
      y * CELL_SIZE,
      CELL_SIZE,
      CELL_SIZE
    )
  })
}

// 绘制所有蛇
function drawSnakes () {
  console.log('绘制蛇数量:', players.length)

  if (!players || players.length === 0) {
    console.log('没有玩家可绘制')
    return
  }

  players.forEach((player, index) => {
    console.log(`绘制玩家 ${index}:`, player)

    if (!player) {
      console.error('无效的玩家对象')
      return
    }

    if (!player.snake) {
      console.log('玩家没有蛇:', player.id)
      return
    }

    if (!player.snake.body || !Array.isArray(player.snake.body) || player.snake.body.length === 0) {
      console.log('玩家蛇没有身体部分:', player.id)
      return
    }

    const colorIndex = index % SNAKE_COLORS.length
    drawSnake(player.snake, SNAKE_COLORS[colorIndex], player.id === playerId)
  })
}

// 绘制单条蛇
function drawSnake (snake, color, isCurrentPlayer) {
  if (!snake || !snake.body || !Array.isArray(snake.body)) {
    console.error('无效的蛇对象:', snake)
    return
  }

  ctx.fillStyle = color

  // 绘制蛇身
  snake.body.forEach((segment, index) => {
    if (!segment || typeof segment.x === 'undefined' || typeof segment.y === 'undefined') {
      console.error('无效的蛇身体部分:', segment)
      return
    }

    // 确保坐标是数字
    const x = Number(segment.x)
    const y = Number(segment.y)

    if (isNaN(x) || isNaN(y)) {
      console.error('蛇身体坐标无效:', segment)
      return
    }

    // 如果是当前玩家的蛇头，用不同颜色标记
    if (index === 0 && isCurrentPlayer) {
      ctx.fillStyle = '#000'
      ctx.fillRect(
        x * CELL_SIZE,
        y * CELL_SIZE,
        CELL_SIZE,
        CELL_SIZE
      )
      ctx.fillStyle = color
    } else {
      ctx.fillRect(
        x * CELL_SIZE,
        y * CELL_SIZE,
        CELL_SIZE,
        CELL_SIZE
      )
    }
  })
}

// 显示消息
function showMessage (text) {
  const messageElement = document.getElementById('message')
  messageElement.textContent = text
  messageElement.classList.add('show')

  // 清除之前的定时器
  if (messageTimeout) {
    clearTimeout(messageTimeout)
  }

  // 3秒后隐藏消息
  messageTimeout = setTimeout(() => {
    messageElement.classList.remove('show')
  }, 3000)
}

// 页面加载完成后初始化游戏
window.onload = init
window.addEventListener('load', init)

// 添加调试信息
console.log('游戏脚本已加载')