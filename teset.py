import argparse
from git_filter_repo import RepoFilter

def main():
    # 创建解析器
    parser = argparse.ArgumentParser()

    # 添加参数
    parser.add_argument('--path', type=str, required=True, help="Path to the Git repository")

    # 解析参数
    args = parser.parse_args(['--path', '.'])  # '.' 表示当前目录，你可以根据需要修改为具体路径# 手动添加缺失的回调属性
    args.filename_callback = None
    args.message_callback = None
    args.name_callback = None
    args.email_callback = None
    args.refname_callback = None
    args.blob_callback = None
    args.commit_callback = 'commit_callback.py'  # 回调函数文件路径
    args.tag_callback = None
    args.reset_callback = None
    args.done_callback = None  # 初始化 RepoFilter，并传入 args
    repo_filter = RepoFilter(args=args)

    # 运行过滤操作
    repo_filter.run()

if __name__ == "__main__":
    main()
