import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const publicPaths = ['/login', '/callback', '/'];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // 静的アセット・API ルートはスキップ
  if (pathname.startsWith('/_next') || pathname.startsWith('/api') || /\.(?:png|jpg|jpeg|gif|svg|ico|webp|woff2?)$/i.test(pathname)) {
    return NextResponse.next();
  }

  // ログイン済みユーザーが /login にアクセス → /dashboard へリダイレクト
  if (pathname === '/login') {
    const token = request.cookies.get('tb_jwt');
    if (token) {
      return NextResponse.redirect(new URL('/dashboard', request.url));
    }
  }

  // パブリックパスはスキップ（完全一致 or サブパス一致）
  if (publicPaths.some((p) => pathname === p || (p !== '/' && pathname.startsWith(p + '/')))) {
    return NextResponse.next();
  }

  // JWT Cookie の存在チェック（検証はバックエンドに委譲）
  const token = request.cookies.get('tb_jwt');
  if (!token) {
    return NextResponse.redirect(new URL('/', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
