#!/usr/bin/env python3
"""
Cal1604 应用程序图标生成器
生成符合项目设计规范的 SVG 和 PNG 图标
"""

import os
import sys

def generate_svg_icon():
    """生成 SVG 图标"""
    # 配色方案来自 DESIGN.md
    bg_color = "#1e1e1e"  # 主背景色
    bg_secondary = "#252526"  # 次级背景
    accent_gold = "#ffd700"  # 明金强调色
    accent_gold_dark = "#d4a800"  # 深金色
    text_color = "#d4d4d4"  # 主要文字色
    
    svg_content = f'''<?xml version="1.0" encoding="UTF-8"?>
<svg width="1024" height="1024" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <!-- 背景渐变 -->
    <radialGradient id="bgGradient" cx="50%" cy="50%" r="50%" fx="50%" fy="50%">
      <stop offset="0%" style="stop-color:#2d2d2d;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#1e1e1e;stop-opacity:1" />
    </radialGradient>
    
    <!-- 金色渐变 -->
    <linearGradient id="goldGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#ffed4a;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#ffd700;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#d4a800;stop-opacity:1" />
    </linearGradient>
    
    <!-- 金色发光效果 -->
    <filter id="goldGlow" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur stdDeviation="4" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
    
    <!-- 阴影效果 -->
    <filter id="dropShadow" x="-20%" y="-20%" width="140%" height="140%">
      <feGaussianBlur in="SourceAlpha" stdDeviation="8"/>
      <feOffset dx="4" dy="4" result="offsetblur"/>
      <feComponentTransfer>
        <feFuncA type="linear" slope="0.5"/>
      </feComponentTransfer>
      <feMerge>
        <feMergeNode/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>
  
  <!-- 背景圆角矩形 -->
  <rect x="32" y="32" width="960" height="960" rx="160" ry="160" fill="url(#bgGradient)"/>
  
  <!-- 外边框 -->
  <rect x="48" y="48" width="928" height="928" rx="144" ry="144" 
        fill="none" stroke="#333333" stroke-width="4"/>
  
  <!-- 内部装饰圆环 -->
  <circle cx="512" cy="512" r="380" fill="none" stroke="#333333" stroke-width="2"/>
  <circle cx="512" cy="512" r="360" fill="none" stroke="#444444" stroke-width="1"/>
  
  <!-- 压力表主体 - 外圈 -->
  <circle cx="512" cy="480" r="280" fill="none" stroke="url(#goldGradient)" stroke-width="8" filter="url(#dropShadow)"/>
  
  <!-- 压力表刻度盘背景 -->
  <circle cx="512" cy="480" r="260" fill="#252526" stroke="#333333" stroke-width="2"/>
  
  <!-- 刻度线 - 主刻度 -->
  <g stroke="url(#goldGradient)" stroke-width="4" stroke-linecap="round">
    <line x1="512" y1="240" x2="512" y2="270" />
    <line x1="752" y1="480" x2="722" y2="480" />
    <line x1="512" y1="720" x2="512" y2="690" />
    <line x1="272" y1="480" x2="302" y2="480" />
    <!-- 45度刻度 -->
    <line x1="682" y1="310" x2="661" y2="331" stroke-width="3"/>
    <line x1="682" y1="650" x2="661" y2="629" stroke-width="3"/>
    <line x1="342" y1="650" x2="363" y2="629" stroke-width="3"/>
    <line x1="342" y1="310" x2="363" y2="331" stroke-width="3"/>
  </g>
  
  <!-- 刻度线 - 次刻度 -->
  <g stroke="#555555" stroke-width="2" stroke-linecap="round">
    <line x1="622" y1="262" x2="612" y2="281" />
    <line x1="700" y1="372" x2="681" y2="382" />
    <line x1="700" y1="588" x2="681" y2="578" />
    <line x1="622" y1="698" x2="612" y2="679" />
    <line x1="402" y1="698" x2="412" y2="679" />
    <line x1="324" y1="588" x2="343" y2="578" />
    <line x1="324" y1="372" x2="343" y2="382" />
    <line x1="402" y1="262" x2="412" y2="281" />
  </g>
  
  <!-- 中心点 -->
  <circle cx="512" cy="480" r="24" fill="url(#goldGradient)" filter="url(#goldGlow)"/>
  <circle cx="512" cy="480" r="12" fill="#1e1e1e"/>
  
  <!-- 指针 - 指向 12 点方向 -->
  <g transform="rotate(0, 512, 480)">
    <polygon points="512,260 502,480 512,500 522,480" fill="url(#goldGradient)" filter="url(#dropShadow)"/>
    <circle cx="512" cy="480" r="8" fill="url(#goldGradient)"/>
  </g>
  
  <!-- 校准标记 - 对勾符号 -->
  <g transform="translate(680, 280)" filter="url(#goldGlow)">
    <path d="M 0 40 L 30 70 L 80 10" fill="none" stroke="url(#goldGradient)" stroke-width="12" stroke-linecap="round" stroke-linejoin="round"/>
  </g>
  
  <!-- 型号文字 1604 -->
  <text x="512" y="840" text-anchor="middle" font-family="Arial, sans-serif" font-size="120" font-weight="bold" fill="url(#goldGradient)" letter-spacing="16" filter="url(#dropShadow)">1604</text>
  
  <!-- CAL 标识 -->
  <text x="512" y="600" text-anchor="middle" font-family="Arial, sans-serif" font-size="48" font-weight="500" fill="#888888" letter-spacing="8">CALIBRATION</text>
  
  <!-- 装饰性圆点 -->
  <circle cx="160" cy="160" r="12" fill="#333333"/>
  <circle cx="864" cy="160" r="12" fill="#333333"/>
  <circle cx="160" cy="864" r="12" fill="#333333"/>
  <circle cx="864" cy="864" r="12" fill="#333333"/>
</svg>'''
    
    return svg_content

def save_svg(filepath, content):
    """保存 SVG 文件"""
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"SVG 图标已保存: {filepath}")

def convert_svg_to_png(svg_path, png_path, sizes=None):
    """将 SVG 转换为 PNG"""
    if sizes is None:
        sizes = [256, 512, 1024]
    
    try:
        from cairosvg import svg2png
        
        # 生成主图标
        for size in sizes:
            output_path = png_path.replace('.png', f'_{size}.png')
            svg2png(url=svg_path, write_to=output_path, output_width=size, output_height=size)
            print(f"PNG 图标已生成 ({size}x{size}): {output_path}")
        
        # 生成标准尺寸的图标
        svg2png(url=svg_path, write_to=png_path, output_width=1024, output_height=1024)
        print(f"PNG 图标已保存: {png_path}")
        return True
        
    except ImportError:
        print("警告: 未安装 cairosvg，跳过 PNG 转换")
        print("如需 PNG 格式，请运行: pip install cairosvg")
        return False

def create_ico(svg_path, ico_path):
    """创建 Windows ICO 文件"""
    try:
        from PIL import Image
        import io
        from cairosvg import svg2png
        
        # 读取 SVG 并生成多种尺寸的 PNG
        sizes = [16, 24, 32, 48, 64, 128, 256]
        images = []
        
        for size in sizes:
            png_data = svg2png(url=svg_path, output_width=size, output_height=size)
            img = Image.open(io.BytesIO(png_data))
            if img.mode != 'RGBA':
                img = img.convert('RGBA')
            images.append(img)
        
        # 保存为 ICO
        images[0].save(ico_path, format='ICO', sizes=[(s, s) for s in sizes], append_images=images[1:])
        print(f"ICO 图标已保存: {ico_path}")
        return True
        
    except ImportError as e:
        print(f"警告: 创建 ICO 需要 PIL 和 cairosvg: {e}")
        return False

def main():
    """主函数"""
    # 项目根目录
    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    build_dir = os.path.join(project_root, 'build')
    
    # 确保 build 目录存在
    os.makedirs(build_dir, exist_ok=True)
    os.makedirs(os.path.join(build_dir, 'windows'), exist_ok=True)
    
    # 文件路径
    svg_path = os.path.join(build_dir, 'appicon.svg')
    png_path = os.path.join(build_dir, 'appicon.png')
    ico_path = os.path.join(build_dir, 'windows', 'icon.ico')
    
    print("=" * 50)
    print("Cal1604 图标生成器")
    print("=" * 50)
    
    # 生成 SVG
    print("\n1. 生成 SVG 图标...")
    svg_content = generate_svg_icon()
    save_svg(svg_path, svg_content)
    
    # 转换为 PNG
    print("\n2. 转换 PNG 图标...")
    png_success = convert_svg_to_png(svg_path, png_path)
    
    # 创建 ICO
    print("\n3. 创建 Windows ICO 图标...")
    ico_success = create_ico(svg_path, ico_path)
    
    print("\n" + "=" * 50)
    print("图标生成完成!")
    print("=" * 50)
    print(f"\n生成的文件:")
    print(f"  - SVG: {svg_path}")
    if png_success:
        print(f"  - PNG: {png_path}")
    if ico_success:
        print(f"  - ICO: {ico_path}")
    
    print("\n图标设计说明:")
    print("  - 深色背景 (#1e1e1e) 符合项目设计规范")
    print("  - 金色压力表 (#ffd700) 体现精确校准主题")
    print("  - 1604 型号标识突出产品特征")
    print("  - 现代扁平设计配合微妙渐变")

if __name__ == '__main__':
    main()
