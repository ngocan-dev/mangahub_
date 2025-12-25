import { getChapterById, type ChapterDetail } from "@/service/api";

async function getById(id: number): Promise<ChapterDetail> {
  return getChapterById(id);
}

export const chapterService = {
  getById,
};

export type { ChapterDetail };
