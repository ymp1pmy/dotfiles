local wezterm = require("wezterm")

local act = wezterm.action

local wsl_domains = wezterm.default_wsl_domains()
for _, dom in ipairs(wsl_domains) do
    dom.default_cwd = "~"
end

-- 利用可能な WSL ディストリの中から使いたいものを優先順で選ぶ。
-- 環境ごとに導入済みの Ubuntu バージョンが違っても動くよう、
-- ベタ書きせず「あるものの中から」フォールバックする。
local function pick_default_domain(domains, preferred)
    local available = {}
    for _, dom in ipairs(domains) do
        available[dom.name] = true
    end

    for _, name in ipairs(preferred) do
        if available[name] then
            return name
        end
    end

    -- 優先候補がどれも無ければ、先頭の WSL ドメインにフォールバック。
    return domains[1] and domains[1].name or nil
end

local function split(str, ptr)
    local splitted = {}
    for token in string.gmatch(str, string.format("[^%s]+", ptr)) do
        table.insert(splitted, token)
    end

    return splitted
end

local function get_current_working_dir(tab)
    local current_dir = tab.active_pane and tab.active_pane.current_working_dir or { file_path = "" }
    local home_dir = string.format("file://%s", os.getenv("HOME"))
    local path = split(current_dir.file_path, "/")

    return current_dir == home_dir and "." or path[#path]
end

wezterm.on("format-tab-title", function(tab, tabs, panes, config, hover, max_width)
    local has_unseen_output = false
    if not tab.is_active then
        for _, pane in ipairs(tab.panes) do
            if pane.has_unseen_output then
                has_unseen_output = true
                break
            end
        end
    end

    local cwd = wezterm.format({
        { Attribute = { Intensity = "Bold" } },
        { Text = get_current_working_dir(tab) },
    })

    local title = string.format(" %s: %s ", tab.tab_index, cwd)

    if has_unseen_output then
        return {
            { Foreground = { Color = "#28719c" } },
            { Text = title },
        }
    end

    return {
        { Text = title },
    }
end)

-- ベル (BEL) を受け取ったら OS のトースト通知も出す。
-- Claude Code のフックが「処理完了」「入力待ち」で BEL を撃ってくるので、
-- 別ウィンドウを見ていても気づけるようにする。
-- WSL 越し / tmux 越しでも BEL はそのまま伝わってくる。
wezterm.on("bell", function(window, pane)
    window:toast_notification("WezTerm", string.format("%s が呼んでいます", pane:get_title()), nil, 4000)
end)

local config = {}

config.default_domain = pick_default_domain(wsl_domains, {
    "WSL:Ubuntu-26.04",
    "WSL:Ubuntu-24.04",
})
config.wsl_domains = wsl_domains

config.color_scheme = "Kanagawa (Gogh)"
config.colors = {
    foreground = "#dcd7ba",
    background = "#181616",

    cursor_bg = "#c8c093",
    cursor_fg = "#2d4f67",
    cursor_border = "#c8c093",

    selection_fg = "#c8c093",
    selection_bg = "#2d4f67",

    scrollbar_thumb = "#16161d",
    split = "#c8c093",

    -- visual_bell のフラッシュ色 (Kanagawa の選択色に合わせる)
    visual_bell = "#2d4f67",

    ansi = { "#090618", "#c34043", "#76946a", "#c0a36e", "#7e9cd8", "#957fb8", "#6a9589", "#c8c093" },
    brights = { "#727169", "#e82424", "#98bb6c", "#e6c384", "#7fb4ca", "#938aa9", "#7aa89f", "#dcd7ba" },
    indexed = { [16] = "#ffa066", [17] = "#ff5d62" },
}

config.font = wezterm.font_with_fallback({
    { family = "UDEV Gothic 35NFLG", weight = 600 },
    { family = "SauceCodePro NFM" },
})
config.font_size = 9
config.line_height = 1.3

-- SEE: https://wezfurlong.org/wezterm/config/lua/config/canonicalize_pasted_newlines.html
config.canonicalize_pasted_newlines = "LineFeed"
config.window_decorations = "RESIZE"
-- 1920x1080 モニタで画面からはみ出さず快適に使えるサイズ。
-- line_height 1.3 の分だけ 1 行が高いので、rows は控えめにしておく。
config.initial_rows = 40
config.initial_cols = 150
config.window_padding = {
    left = "2cell",
    right = "2cell",
    top = "0cell",
    bottom = "0cell",
}

config.default_cursor_style = "BlinkingBlock"
config.cursor_blink_rate = 500
config.cursor_blink_ease_in = "Constant"
config.cursor_blink_ease_out = "Constant"

-- ベルは「音 + 画面フラッシュ + トースト通知」の三段構え。
-- 端末側 (Claude Code のフックなど) が \a を吐けばここで拾う。
config.audible_bell = "SystemBeep"
config.visual_bell = {
    fade_in_duration_ms = 75,
    fade_in_function = "EaseIn",
    fade_out_duration_ms = 150,
    fade_out_function = "EaseOut",
    target = "BackgroundColor",
}

config.inactive_pane_hsb = {
    saturation = 1.0,
    brightness = 1.0,
}

config.tab_bar_at_bottom = true
config.use_fancy_tab_bar = false
config.colors.tab_bar = {
    background = "#181616",
    active_tab = {
        bg_color = "#2d4f67",
        fg_color = "#dcd7ba",
        intensity = "Bold",
    },
    inactive_tab = {
        bg_color = "#181616",
        fg_color = "#727169",
    },
    inactive_tab_hover = {
        bg_color = "#1f1f28",
        fg_color = "#dcd7ba",
    },
    new_tab = {
        bg_color = "#181616",
        fg_color = "#727169",
    },
}

return config
